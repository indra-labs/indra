// Package node maintains information about peers on the network and associated
// connection sessions.
package node

import (
	"fmt"
	"net/netip"
	"sync"
	"time"

	"github.com/indra-labs/indra"
	"github.com/indra-labs/indra/pkg/ifc"
	"github.com/indra-labs/indra/pkg/key/prv"
	"github.com/indra-labs/indra/pkg/key/pub"
	log2 "github.com/indra-labs/indra/pkg/log"
	"github.com/indra-labs/indra/pkg/nonce"
	"github.com/indra-labs/indra/pkg/slice"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

// Node is a representation of a messaging counterparty. The netip.AddrPort can
// be nil for the case of a client node that is not in a direct open connection,
// or for the special node in a client. For this reason all nodes are assigned
// an ID and will normally be handled by this except when the netip.AddrPort is
// known via the packet sender address.
type Node struct {
	nonce.ID
	sync.Mutex
	Addr          string
	AddrPort      *netip.AddrPort
	IdentityPub   *pub.Key
	IdentityBytes pub.Bytes
	IdentityPrv   *prv.Key
	PingCount     int
	LastSeen      time.Time
	*Circuit
	Sessions
	Services
	ifc.Transport
}

// New creates a new Node. netip.AddrPort is optional if the counterparty is not
// in direct connection.
func New(addr *netip.AddrPort, idPub *pub.Key, idPrv *prv.Key,
	tpt ifc.Transport) (n *Node, id nonce.ID) {

	id = nonce.NewID()
	n = &Node{
		ID:            id,
		Addr:          addr.String(),
		AddrPort:      addr,
		Transport:     tpt,
		IdentityPub:   idPub,
		IdentityBytes: idPub.ToBytes(),
		IdentityPrv:   idPrv,
	}
	n.Sessions = append(n.Sessions, NewSession(id, n, 0))
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
			log.I.Ln("receivefrom")
			b = n.Services[i].Receive()
			return
		}
	}
	return
}
