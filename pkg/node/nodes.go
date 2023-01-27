package node

import (
	"fmt"
	"net/netip"

	"github.com/indra-labs/indra/pkg/crypto/nonce"
)

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
