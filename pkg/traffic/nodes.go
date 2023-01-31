package traffic

import (
	"fmt"
	"net/netip"

	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
)

// NodesLen returns the length of a Nodes.
func (sm *SessionManager) NodesLen() int {
	sm.Lock()
	defer sm.Unlock()
	return len(sm.nodes)
}

// GetLocalNode returns the engine's local Node.
func (sm *SessionManager) GetLocalNode() *Node { return sm.nodes[0] }

// AddNodes adds a Node to a Nodes.
func (sm *SessionManager) AddNodes(nn ...*Node) {
	sm.Lock()
	defer sm.Unlock()
	sm.nodes = append(sm.nodes, nn...)
}

// FindNodeByID searches for a Node by ID.
func (sm *SessionManager) FindNodeByID(i nonce.ID) (no *Node) {
	sm.Lock()
	defer sm.Unlock()
	for _, nn := range sm.nodes {
		if nn.ID == i {
			no = nn
			break
		}
	}
	return
}

// FindNodeByAddrPort searches for a Node by netip.AddrPort.
func (sm *SessionManager) FindNodeByAddrPort(id *netip.AddrPort) (no *Node) {
	sm.Lock()
	defer sm.Unlock()
	for _, nn := range sm.nodes {
		if nn.AddrPort.String() == id.String() {
			no = nn
			break
		}
	}
	return
}

// DeleteNodeByID deletes a node identified by an ID.
func (sm *SessionManager) DeleteNodeByID(ii nonce.ID) (e error) {
	sm.Lock()
	defer sm.Unlock()
	e = fmt.Errorf("id %x not found", ii)
	for i := range sm.nodes {
		if sm.nodes[i].ID == ii {
			sm.nodes = append(sm.nodes[:i], sm.nodes[i+1:]...)
			return
		}
	}
	return
}

// DeleteNodeByAddrPort deletes a node identified by a netip.AddrPort.
func (sm *SessionManager) DeleteNodeByAddrPort(ip *netip.AddrPort) (e error) {
	sm.Lock()
	defer sm.Unlock()
	e = fmt.Errorf("node with ip %v not found", ip)
	for i := range sm.nodes {
		if sm.nodes[i].AddrPort.String() == ip.String() {
			sm.nodes = append(sm.nodes[:i], sm.nodes[i+1:]...)
			e = nil
			break
		}
	}
	return
}

// ForEachNode runs a function over the slice of nodes with the mutex locked,
// and terminates when the function returns true.
func (sm *SessionManager) ForEachNode(fn func(n *Node) bool) {
	sm.Lock()
	defer sm.Unlock()
	for i := range sm.nodes {
		if fn(sm.nodes[i]) {
			return
		}
	}
}
