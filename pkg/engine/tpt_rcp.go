package engine

import (
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/multiformats/go-multiaddr"
	
	"git-indra.lan/indra-labs/indra"
	"git-indra.lan/indra-labs/indra/pkg/crypto"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

const (
	RCP_MTU         = 1440
	RCP_ID          = "/indra/relay/" + indra.SemVer
	RCP_ServiceName = "indra.relay"
)

type RCPListener struct {
	C    ByteChan
	mtu  int
	Host host.Host
}

func NewRCPListener(multiAddr string, prv *crypto.Prv, mtu,
	bufs int) (c *RCPListener, e error) {
	
	c = &RCPListener{C: make(ByteChan, bufs), mtu: mtu}
	var ma multiaddr.Multiaddr
	if ma, e = multiaddr.NewMultiaddr(multiAddr); fails(e) {
		return
	}
	if c.Host, e = libp2p.New(
		libp2p.Identity(prv),
		libp2p.ListenAddrs(ma),
		libp2p.NoSecurity,
		libp2p.EnableHolePunching(),
	); fails(e) {
		return
	}
	c.Host.SetStreamHandler(RCP_ID, c.handle)
	return
}

func (c *RCPListener) handle(s network.Stream) {
	id := s.ID()
	log.D.S("id", id)
	b := slice.NewBytes(RCP_MTU)
	var e error
	var n int
	if n, e = s.Read(b); fails(e) {
		return
	}
	c.C <- b[:n]
}
