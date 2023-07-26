package transport

import (
	"bufio"
	"context"
	"fmt"
	"github.com/indra-labs/indra/pkg/engine/protocols"
	"github.com/indra-labs/indra/pkg/engine/transport/pstoreds"
	badger "github.com/ipfs/go-ds-badger"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	"github.com/multiformats/go-multiaddr"
	"sync"

	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/engine/tpt"
	"github.com/indra-labs/indra/pkg/interrupt"
	"github.com/indra-labs/indra/pkg/util/qu"
	"github.com/indra-labs/indra/pkg/util/slice"
)

// concurrent safe accessors:

type (
	// Listener is a libp2p connected network transport with a DHT and peer
	// information gossip exchange system that carries the advertisements for
	// services and relays providing access to them.
	Listener struct {

		// The DHT is used by peer discovery and peer information gossip to
		// provide information to clients to enable them to set up sessions and
		// then send traffic through them.
		DHT *dht.IpfsDHT

		// MTU is the size of the packets that should be used by the dispatcher,
		// which should be the same as the Path MTU between the node and its
		// routeable internet address.
		MTU int

		// Host is the libp2p peer management system and provides access to all
		// the connectivity and related services provided by libp2p.
		Host host.Host

		// connections are the list of currently open connections being handled
		// by the Listener.
		connections map[string]*Conn

		// newConns is a channel that new inbound connections are dispatched to
		// for a handler to be assigned.
		newConns chan *Conn

		// Keys are this node's identity keys, used for identification and
		// authentication of peer advertisements.
		*crypto.Keys

		// Context here allows listener processes to be signalled to shut down.
		context.Context

		// This shuts down this and all child goroutines.
		context.CancelFunc

		// Mutex because any number of threads can be handling Listener work.
		sync.Mutex
	}

	// Conn is a wrapper around the bidirectional connection established between
	// two peers via the libp2p API.

)

// NewListener creates a new Listener with the given parameters.
func NewListener(rendezvous, multiAddr []string, keys *crypto.Keys, store *badger.Datastore, closer func(), ctx context.Context, mtu int, cancel context.CancelFunc) (c *Listener, e error) {

	c = &Listener{
		Keys:        keys,
		MTU:         mtu,
		connections: make(map[string]*Conn),
		newConns:    make(chan *Conn, ConnBufs),
		Context:     ctx,
		CancelFunc:  cancel,
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
	interrupt.AddHandler(closer)
	var st peerstore.Peerstore
	st, e = pstoreds.NewPeerstore(ctx, store, pstoreds.DefaultOpts())
	if c.Host, e = libp2p.New(
		libp2p.Identity(keys.Prv),
		libp2p.UserAgent(DefaultUserAgent),
		libp2p.ListenAddrs(ma...),
		libp2p.EnableHolePunching(),
		// libp2p.Transport(libp2pquic.NewTransport),
		libp2p.Transport(tcp.NewTCPTransport),
		// libp2p.Transport(websocket.New),
		// libp2p.Security(libp2ptls.ID, libp2ptls.New),
		libp2p.Security(noise.ID, noise.New),
		// libp2p.NoSecurity,
		libp2p.Peerstore(st),
	); fails(e) {
		return
	}
	interrupt.AddHandler(closer)
	if c.DHT, e = NewDHT(ctx, c.Host, rdv); fails(e) {
		return
	}
	go c.Discover(ctx, c.Host, c.DHT, rdv)
	c.Host.SetStreamHandler(IndraLibP2PID, c.handle)
	return
}

func (l *Listener) ProtocolsAvailable() (p protocols.NetworkProtocols) {
	if l == nil || l.Host == nil {
		return protocols.IP4 | protocols.IP6
	}
	for _, v := range l.Host.Addrs() {
		if _, e := v.ValueForProtocol(multiaddr.P_IP4); e == nil {
			p &= protocols.IP4
		}
		if _, e := v.ValueForProtocol(multiaddr.P_IP4); e == nil {
			p &= protocols.IP6
		}
	}
	return
}

// Accept waits on inbound connections and hands them out to listeners on the
// returned channel.
func (l *Listener) Accept() <-chan *Conn { return l.newConns }

// AddConn adds a new connection, including dispatching it to any callers of
// GetNewConnChan.
func (l *Listener) AddConn(d *Conn) {
	l.newConns <- d
	l.Lock()
	l.connections[d.MultiAddr.String()] = d
	l.Unlock()
}

// DelConn closes and removes the connection from the Listener.
func (l *Listener) DelConn(d *Conn) {
	l.Lock()
	l.connections[d.MultiAddr.String()].Q()
	delete(l.connections, d.MultiAddr.String())
	l.Unlock()
}

// Dial creates a new Conn as an outbound connection.
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
	hostAddress := GetHostOnlyFirstMultiaddr(d.Host)
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

// FindConn searches a Listener's collection of connections for the one matching
// a given multiaddr string.
func (l *Listener) FindConn(multiAddr string) (d *Conn) {
	l.Lock()
	var ok bool
	if d, ok = l.connections[multiAddr]; ok {
	}
	l.Unlock()
	return
}

// GetConnRecv searches a Listener's collection of connections for the one
// matching a given multiaddr string, and returns the transport for receiving
// messages.
func (l *Listener) GetConnRecv(multiAddr string) (recv tpt.Transport) {
	l.Lock()
	if _, ok := l.connections[multiAddr]; ok {
		recv = l.connections[multiAddr].Transport.Receiver
	}
	l.Unlock()
	return
}

// GetConnSend searches a Listener's collection of connections for the one
// matching a given multiaddr string, and returns the transport for sending
// messages.
func (l *Listener) GetConnSend(multiAddr string) (send tpt.Transport) {
	l.Lock()
	if _, ok := l.connections[multiAddr]; ok {
		send = l.connections[multiAddr].Transport.Sender
	}
	l.Unlock()
	return
}

// SetMTU sets the packet segment size to be used for future connections.
func (l *Listener) SetMTU(mtu int) {
	l.Lock()
	l.MTU = mtu
	l.Unlock()
}

// handle processes incoming messages from the network, dispatching to the
// internal channels for processing.
func (l *Listener) handle(s network.Stream) {
	for {
		b := slice.NewBytes(l.MTU)
		var e error
		var n int
		if n, e = s.Read(b); fails(e) {
			return
		}
		log.T.S(blue(GetHostOnlyFirstMultiaddr(l.
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
