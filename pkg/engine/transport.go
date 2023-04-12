package engine

import (
	"bufio"
	"context"
	"fmt"
	"sync"
	"time"
	
	"github.com/cybriq/qu"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/multiformats/go-multiaddr"
	
	"git-indra.lan/indra-labs/indra"
	"git-indra.lan/indra-labs/indra/pkg/crypto"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

const (
	ConnBufs         = 64
	IndraLibP2PID    = "/indra/relay/" + indra.SemVer
	IndraServiceName = "org.indra.relay"
)

type Listener struct {
	MTU         int
	Host        host.Host
	connections map[string]*Conn
	newConns    chan *Conn
	context.Context
	sync.Mutex
}

func NewListener(rendezvous, multiAddr string,
	prv *crypto.Prv, ctx context.Context, mtu int) (c *Listener, e error) {
	
	c = &Listener{
		MTU:         mtu,
		connections: make(map[string]*Conn),
		newConns:    make(chan *Conn, ConnBufs),
		Context:     ctx,
	}
	var ma multiaddr.Multiaddr
	if ma, e = multiaddr.NewMultiaddr(multiAddr); fails(e) {
		return
	}
	rdv := make([]multiaddr.Multiaddr, 1)
	if rendezvous != "" {
		if rdv[0], e = multiaddr.NewMultiaddr(rendezvous); fails(e) {
			return
		}
	} else {
		rdv = nil
	}
	if c.Host, e = libp2p.New(
		libp2p.Identity(prv),
		libp2p.ListenAddrs(ma),
		libp2p.NoSecurity,
		libp2p.EnableHolePunching(),
	); fails(e) {
		return
	}
	var d *dht.IpfsDHT
	if d, e = NewDHT(ctx, c.Host, rdv); fails(e) {
		return
	}
	log.D.Ln("listener", getHostAddress(c.Host))
	go Discover(ctx, c.Host, d, rendezvous)
	c.Host.SetStreamHandler(IndraLibP2PID, c.handle)
	return
}
func (l *Listener) SetMTU(mtu int) {
	l.Lock()
	l.MTU = mtu
	l.Unlock()
}

func (l *Listener) handle(s network.Stream) {
	for {
		b := slice.NewBytes(l.MTU)
		var e error
		var n int
		if n, e = s.Read(b); fails(e) {
			return
		}
		log.D.S(getHostAddress(l.Host)+" read from listener", b[:n].ToBytes())
		id := s.Conn().RemotePeer()
		ai := l.Host.Peerstore().PeerInfo(id)
		aid := ai.ID.String()
		hostAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/p2p/%s", aid))
		ha := ai.Addrs[0].Encapsulate(hostAddr)
		as := ha.String()
		nc := l.Dial(as)
		if l.FindConn(as) == nil {
			l.AddConn(nc)
		}
		nc.Recv <- b[:n]
	}
}

func (l *Listener) Accept() <-chan *Conn { return l.newConns }

func (l *Listener) FindConn(multiAddr string) (d *Conn) {
	l.Lock()
	var ok bool
	if d, ok = l.connections[multiAddr]; ok {
	}
	l.Unlock()
	return
}

func (l *Listener) AddConn(d *Conn) {
	l.newConns <- d
	l.Lock()
	l.connections[d.MultiAddr.String()] = d
	l.Unlock()
}

func (l *Listener) DelConn(d *Conn) {
	l.Lock()
	l.connections[d.MultiAddr.String()].Q()
	delete(l.connections, d.MultiAddr.String())
	l.Unlock()
}

func (l *Listener) GetConnSend(multiAddr string) (send ByteChan) {
	l.Lock()
	if _, ok := l.connections[multiAddr]; ok {
		send = l.connections[multiAddr].Send
	}
	l.Unlock()
	return
}

func (l *Listener) GetConnRecv(multiAddr string) (recv ByteChan) {
	l.Lock()
	if _, ok := l.connections[multiAddr]; ok {
		recv = l.connections[multiAddr].Recv
	}
	l.Unlock()
	return
}

type DuplexByteChan struct {
	Recv, Send ByteChan
}

func NewDuplexByteChan(bufs int) *DuplexByteChan {
	return &DuplexByteChan{make(ByteChan, bufs), make(ByteChan, bufs)}
}

type Conn struct {
	MTU       int
	MultiAddr multiaddr.Multiaddr
	Host      host.Host
	rw        *bufio.ReadWriter
	*DuplexByteChan
	qu.C
}

func getHostAddress(ha host.Host) string {
	hostAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/p2p/%s",
		ha.ID().String()))
	
	// Now we can build a full multiaddress to reach this host by encapsulating
	// both addresses:
	addr := ha.Addrs()[0]
	return addr.Encapsulate(hostAddr).String()
}

func (l *Listener) Dial(multiAddr string) (d *Conn) {
	
	var e error
	var ma multiaddr.Multiaddr
	if ma, e = multiaddr.NewMultiaddr(multiAddr); fails(e) {
		return
	}
	var info *peer.AddrInfo
	if info, e = peer.AddrInfoFromP2pAddr(ma); fails(e) {
		return
	}
	l.Host.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)
	var s network.Stream
	if s, e = l.Host.NewStream(context.Background(), info.ID, IndraLibP2PID); fails(e) {
		return
	}
	d = &Conn{
		MTU:            l.MTU,
		MultiAddr:      ma,
		Host:           l.Host,
		DuplexByteChan: NewDuplexByteChan(ConnBufs),
		rw:             bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s)),
		C:              qu.T(),
	}
	l.Lock()
	l.connections[multiAddr] = d
	l.Unlock()
	hostAddress := getHostAddress(d.Host)
	go func() {
		var e error
		for {
			log.D.Ln("sender", hostAddress, "ready")
			select {
			case <-d.C:
				return
			case b := <-d.Send:
				log.D.S(hostAddress+" sending to "+d.MultiAddr.String(),
					b.ToBytes())
				if _, e = d.rw.Write(b); fails(e) {
					continue
				}
				if e = d.rw.Flush(); fails(e) {
					continue
				}
				log.D.Ln(hostAddress, "sent")
			}
		}
	}()
	return
}

func NewDHT(ctx context.Context, host host.Host,
	bootstrapPeers []multiaddr.Multiaddr) (*dht.IpfsDHT, error) {
	
	var options []dht.Option
	if len(bootstrapPeers) == 0 {
		options = append(options, dht.Mode(dht.ModeServer))
	}
	kdht, err := dht.New(ctx, host, options...)
	if err != nil {
		return nil, err
	}
	if err = kdht.Bootstrap(ctx); err != nil {
		return nil, err
	}
	var wg sync.WaitGroup
	for _, peerAddr := range bootstrapPeers {
		peerinfo, _ := peer.AddrInfoFromP2pAddr(peerAddr)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if e := host.Connect(ctx, *peerinfo); fails(e) {
				log.E.F("Error while connecting to node %q: %-v", peerinfo, e)
			} else {
				log.I.F("%s: Connection established with bootstrap node: %s",
					getHostAddress(host), *peerinfo)
			}
		}()
	}
	wg.Wait()
	return kdht, nil
}

func Discover(ctx context.Context, h host.Host, dht *dht.IpfsDHT, rendezvous string) {
	var routingDiscovery = routing.NewRoutingDiscovery(dht)
	var e error
	var peers <-chan peer.AddrInfo
	_, e = routingDiscovery.Advertise(ctx, rendezvous)
	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			peers, e = routingDiscovery.FindPeers(ctx, rendezvous)
			if fails(e) {
				return
			}
			for p := range peers {
				if p.ID == h.ID() {
					continue
				}
				if h.Network().Connectedness(p.ID) != network.Connected {
					_, e = h.Network().DialPeer(ctx, p.ID)
					fmt.Printf("Connected to peer %s\n", p.ID.String())
					if e != nil {
						continue
					}
				}
			}
		}
	}
}
