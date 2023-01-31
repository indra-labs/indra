// Package traffic maintains information about peers on the network and
// associated connection sessions.
package traffic

import (
	"fmt"
	"net/netip"

	"git-indra.lan/indra-labs/lnd/lnd/lnwire"

	"git-indra.lan/indra-labs/indra"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/ring"
	"git-indra.lan/indra-labs/indra/pkg/service"
	"git-indra.lan/indra-labs/indra/pkg/types"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

// Node is a representation of a messaging counterparty.
type Node struct {
	nonce.ID
	AddrPort      *netip.AddrPort
	IdentityPub   *pub.Key
	IdentityBytes pub.Bytes
	IdentityPrv   *prv.Key
	RelayRate     lnwire.MilliSatoshi // Base relay price/Mb.
	Services      service.Services    // Services offered by this peer.
	Load          *ring.BufferLoad    // Relay load.
	Latency       *ring.BufferLatency // Latency to peer.
	Failure       *ring.BufferFailure // Times of tx failure.
	types.Transport
}

// DefaultSampleBufferSize defines the number of samples for the Load, Latency
// and Failure ring buffers.
const DefaultSampleBufferSize = 64

// New creates a new Node. A netip.AddrPort is optional if the counterparty is
// not in direct connection. Also, the IdentityPrv node private key can be nil,
// as only the node embedded in a client and not the peer node list has one
// available. The Node for a client's self should use true in the local
// parameter to not initialise the peer state ring buffers as it won't use them.
func New(addr *netip.AddrPort, idPub *pub.Key, idPrv *prv.Key,
	tpt types.Transport, relayRate lnwire.MilliSatoshi,
	local bool) (n *Node, id nonce.ID) {

	id = nonce.NewID()
	n = &Node{
		ID:            id,
		AddrPort:      addr,
		IdentityPub:   idPub,
		IdentityBytes: idPub.ToBytes(),
		IdentityPrv:   idPrv,
		RelayRate:     relayRate,
		Transport:     tpt,
	}
	if !local {
		// These ring buffers are needed to evaluate these metrics for remote
		// peers only.
		n.Load = ring.NewBufferLoad(DefaultSampleBufferSize)
		n.Latency = ring.NewBufferLatency(DefaultSampleBufferSize)
		n.Failure = ring.NewBufferFailure(DefaultSampleBufferSize)
	}
	return
}

// SendTo delivers a message to a service identified by its port.
func (n *Node) SendTo(port uint16, b slice.Bytes) (e error) {
	e = fmt.Errorf("port not registered %d", port)
	for i := range n.Services {
		if n.Services[i].Port == port {
			n.Services[i].Send(b)
			e = nil
			return
		}
	}
	return
}

// ReceiveFrom returns the channel that receives messages for a given port.
func (n *Node) ReceiveFrom(port uint16) (b <-chan slice.Bytes) {
	for i := range n.Services {
		if n.Services[i].Port == port {
			log.T.Ln("receive from")
			b = n.Services[i].Receive()
			return
		}
	}
	return
}
