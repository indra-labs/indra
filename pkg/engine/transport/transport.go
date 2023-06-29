package transport

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/gookit/color"
	"github.com/ipfs/go-datastore"
	badger "github.com/ipfs/go-ds-badger"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/p2p/host/peerstore/pstoreds"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	"github.com/multiformats/go-multiaddr"
	"sync"

	"github.com/indra-labs/indra"
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/engine/tpt"
	"github.com/indra-labs/indra/pkg/interrupt"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/qu"
	"github.com/indra-labs/indra/pkg/util/slice"
)

const (
	// LocalhostZeroIPv4TCP is the default localhost to bind to any address. Used in
	// tests.
	LocalhostZeroIPv4TCP = "/ip4/127.0.0.1/tcp/0"

	// LocalhostZeroIPv4QUIC - Don't use. Buffer problems on linux and fails on CI.
	// LocalhostZeroIPv4QUIC = "/ip4/127.0.0.1/udp/0/quic"

	// DefaultMTU is the default maximum size for a packet.
	DefaultMTU = 1382

	// ConnBufs is the number of buffers to use in message dispatch channels.
	ConnBufs = 8192

	// IndraLibP2PID is the indra protocol identifier.
	IndraLibP2PID = "/indra/relay/" + indra.SemVer
)

var (
	userAgent = "/indra:" + indra.SemVer + "/"

	blue  = color.Blue.Sprint
	log   = log2.GetLogger()
	fails = log.E.Chk
)

// SetUserAgent changes the user agent. Note that this will only have an effect
// before a new listener is created.
func SetUserAgent(s string) {
	userAgent = "/indra " + indra.SemVer + " " + s + "/"
}

// concurrent safe accessors:

func (c *Conn) GetMTU() int {
	c.Lock()
	defer c.Unlock()
	return c.MTU
}

func (c *Conn) GetRecv() tpt.Transport { return c.Transport.Receiver }

func (c *Conn) GetRemoteKey() (remoteKey *crypto.Pub) {
	c.Lock()
	defer c.Unlock()
	return c.RemoteKey
}

func (c *Conn) GetSend() tpt.Transport { return c.Transport.Sender }

func (c *Conn) SetMTU(mtu int) {
	c.Lock()
	c.MTU = mtu
	c.Unlock()
}

func (c *Conn) SetRemoteKey(remoteKey *crypto.Pub) {
	c.Lock()
	c.RemoteKey = remoteKey
	c.Unlock()
}

type (
	Listener struct {
		DHT         *dht.IpfsDHT
		MTU         int
		Host        host.Host
		connections map[string]*Conn
		newConns    chan *Conn
		*crypto.Keys
		context.Context
		sync.Mutex
	}
	Conn struct {
		network.Conn
		MTU       int
		RemoteKey *crypto.Pub
		MultiAddr multiaddr.Multiaddr
		Host      host.Host
		rw        *bufio.ReadWriter
		Transport *DuplexByteChan
		sync.Mutex
		qu.C
	}
)

func GetHostAddress(ha host.Host) string {
	hostAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/p2p/%s",
		ha.ID().String()))
	addr := ha.Addrs()[0]
	return addr.Encapsulate(hostAddr).String()
}
func GetHostOnlyAddress(ha host.Host) string {
	addr := ha.Addrs()[0]
	return addr.String()
}

func (l *Listener) Accept() <-chan *Conn { return l.newConns }

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
	l.Host.Peerstore().AddAddrs(info.ID, info.Addrs,
		peerstore.PermanentAddrTTL)
	var s network.Stream
	if s, e = l.Host.NewStream(context.Background(), info.ID,
		IndraLibP2PID); fails(e) {

		return
	}
	d = &Conn{
		Conn:      s.Conn(),
		MTU:       l.MTU,
		MultiAddr: ma,
		Host:      l.Host,
		Transport: NewDuplexByteChan(ConnBufs),
		rw: bufio.NewReadWriter(bufio.NewReader(s),
			bufio.NewWriter(s)),
		C: qu.T(),
	}
	l.Lock()
	l.connections[multiAddr] = d
	l.Unlock()
	hostAddress := GetHostOnlyAddress(d.Host)
	go func() {
		var e error
		for {
			log.T.Ln(blue(hostAddress), "sender ready")
			select {
			case <-d.C:
				return
			case b := <-d.Transport.Sender.Receive():
				// log.D.S(blue(hostAddress)+" sending to "+
				// 	blue(GetHostOnlyAddress(d.Host)),
				// 	b.ToBytes(),
				// )
				if _, e = d.rw.Write(b); fails(e) {
					continue
				}
				if e = d.rw.Flush(); fails(e) {
					continue
				}
				log.T.Ln(blue(hostAddress), "sent", b.Len(), "bytes")
			}
		}
	}()
	return
}

func (l *Listener) FindConn(multiAddr string) (d *Conn) {
	l.Lock()
	var ok bool
	if d, ok = l.connections[multiAddr]; ok {
	}
	l.Unlock()
	return
}

func (l *Listener) GetConnRecv(multiAddr string) (recv tpt.Transport) {
	l.Lock()
	if _, ok := l.connections[multiAddr]; ok {
		recv = l.connections[multiAddr].Transport.Receiver
	}
	l.Unlock()
	return
}

func (l *Listener) GetConnSend(multiAddr string) (send tpt.Transport) {
	l.Lock()
	if _, ok := l.connections[multiAddr]; ok {
		send = l.connections[multiAddr].Transport.Sender
	}
	l.Unlock()
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
		log.T.S(blue(GetHostOnlyAddress(l.
			Host)) + " read " + fmt.Sprint(n) + " bytes from listener",
		// b[:n].ToBytes(),
		)
		id := s.Conn().RemotePeer()
		ai := l.Host.Peerstore().PeerInfo(id)
		aid := ai.ID.String()
		hostAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/p2p/%s", aid))
		ha := ai.Addrs[0].Encapsulate(hostAddr)
		as := ha.String()
		var nc *Conn
		if nc = l.FindConn(as); nc == nil {
			nc = l.Dial(as)
			if nc == nil {
				log.D.Ln("failed to make connection to", as)
				continue
			}
			l.AddConn(nc)
		}
		nc.Transport.Receiver.Send(b[:n])
	}
}

func NewListener(rendezvous, multiAddr []string, storePath string,
	keys *crypto.Keys, ctx context.Context, mtu int) (c *Listener,
	e error) {

	c = &Listener{
		Keys:        keys,
		MTU:         mtu,
		connections: make(map[string]*Conn),
		newConns:    make(chan *Conn, ConnBufs),
		Context:     ctx,
	}
	var ma []multiaddr.Multiaddr
	for i := range multiAddr {
		var m multiaddr.Multiaddr
		if m, e = multiaddr.NewMultiaddr(multiAddr[i]); fails(e) {
			return
		}
		ma = append(ma, m)
	}
	var rdv []multiaddr.Multiaddr
	if rendezvous != nil || len(rendezvous) > 0 {
		for i := range rendezvous {
			var r multiaddr.Multiaddr
			if r, e = multiaddr.NewMultiaddr(rendezvous[i]); e != nil {
				continue
			}
			rdv = append(rdv, r)
		}
	}
	store, closer := badgerStore(storePath)
	if store == nil {
		return nil, errors.New("could not open database")
	}
	var st peerstore.Peerstore
	st, e = pstoreds.NewPeerstore(ctx, store, pstoreds.DefaultOpts())
	if c.Host, e = libp2p.New(
		libp2p.Identity(keys.Prv),
		libp2p.UserAgent(userAgent),
		libp2p.ListenAddrs(ma...),
		libp2p.EnableHolePunching(),
		//libp2p.Transport(libp2pquic.NewTransport),
		libp2p.Transport(tcp.NewTCPTransport),
		//libp2p.Transport(websocket.New),
		//libp2p.Security(libp2ptls.ID, libp2ptls.New),
		libp2p.Security(noise.ID, noise.New),
		//libp2p.NoSecurity,
		libp2p.Peerstore(st),
	); fails(e) {
		return
	}
	interrupt.AddHandler(closer)
	if c.DHT, e = NewDHT(ctx, c.Host, rdv); fails(e) {
		return
	}
	go Discover(ctx, c.Host, c.DHT, rdv)
	c.Host.SetStreamHandler(IndraLibP2PID, c.handle)
	return
}

func badgerStore(dataPath string) (datastore.Batching, func()) {
	log.T.Ln("dataPath", dataPath)
	store, err := badger.NewDatastore(dataPath, nil)
	if fails(err) {
		return nil, func() {}
	}
	closer := func() {
		store.Close()
	}
	return store, closer
}
