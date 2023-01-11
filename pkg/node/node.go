// Package node maintains information about peers on the network and associated
// connection sessions.
package node

import (
	"fmt"
	"net/netip"
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
// be nil for the case of a client node that is not in a direct open connection.
// For this reason all nodes are assigned an ID and will normally be handled by
// this except when the netip.AddrPort is known via the packet sender address.
type Node struct {
	nonce.ID
	Addr          string
	AddrPort      *netip.AddrPort
	IdentityPub   *pub.Key
	IdentityBytes pub.Bytes
	IdentityPrv   *prv.Key
	PingCount     int
	LastSeen      time.Time
	Services
	ifc.Transport
}

// New creates a new Node. netip.AddrPort is optional if the counterparty is not
// in direct connection.
func New(addr *netip.AddrPort, hdr *pub.Key, hdrPrv *prv.Key,
	tpt ifc.Transport) (n *Node, id nonce.ID) {

	id = nonce.NewID()
	n = &Node{
		ID:            id,
		Addr:          addr.String(),
		AddrPort:      addr,
		Transport:     tpt,
		IdentityPub:   hdr,
		IdentityBytes: hdr.ToBytes(),
		IdentityPrv:   hdrPrv,
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
			log.I.Ln("receivefrom")
			b = n.Services[i].Receive()
			return
		}
	}
	return
}

type Nodes []*Node

// NewNodes creates an empty Nodes
func NewNodes() (n Nodes) { return Nodes{} }

// Len returns the length of a Nodes.
func (n Nodes) Len() int { return len(n) }

// Add a Node to a Nodes.
func (n Nodes) Add(nn *Node) Nodes { return append(n, nn) }

// FindByID searches for a Node by ID.
func (n Nodes) FindByID(i nonce.ID) (no *Node) {
	for _, nn := range n {
		if nn.ID == i {
			no = nn
			break
		}
	}
	return
}

// FindByAddrPort searches for a Node by netip.AddrPort.
func (n Nodes) FindByAddrPort(id *netip.AddrPort) (no *Node) {
	for _, nn := range n {
		if nn.AddrPort.String() == id.String() {
			no = nn
			break
		}
	}
	return
}

// DeleteByID deletes a node identified by an ID.
func (n Nodes) DeleteByID(ii nonce.ID) (nn Nodes, e error) {
	e, nn = fmt.Errorf("id %x not found", ii), n
	for i := range n {
		if n[i].ID == ii {
			return append(n[:i], n[i+1:]...), nil
		}
	}
	return
}

// DeleteByAddrPort deletes a node identified by a netip.AddrPort.
func (n Nodes) DeleteByAddrPort(ip *netip.AddrPort) (nn Nodes, e error) {
	e, nn = fmt.Errorf("node with ip %v not found", ip), n
	for i := range n {
		if n[i].AddrPort.String() == ip.String() {
			nn = append(n[:i], n[i+1:]...)
			e = nil
			break
		}
	}
	return
}
