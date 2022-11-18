// Package node maintains information about peers on the network and associated
// connection sessions.
package node

import (
	"crypto/rand"
	"fmt"
	"net"

	"github.com/Indra-Labs/indra"
	"github.com/Indra-Labs/indra/pkg/session"
	log2 "github.com/cybriq/proc/pkg/log"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

// IDlen is the length of a counterparty ID.
const IDlen = 4

// ID is a randomly generated value used to compactly reference a given node.
// This is only used internally, not given out across the network.
type ID [IDlen]byte

// NewID creates a new random ID.
func NewID() (id ID, e error) {
	var c int
	if c, e = rand.Read(id[:]); check(e) && c != IDlen {
	}
	return
}

// Node is a representation of a messaging counterparty. The net.IP can be nil
// for the case of a client node that is not in a direct open connection. For
// this reason all nodes are assigned an ID and will normally be handled by this
// except when the net.IP is known via the packet sender address.
type Node struct {
	ID
	net.IP
	session.Sessions
}

// New creates a new Node. net.IP is optional if the counterparty is not in
// direct connection.
func New(ip net.IP) (n *Node, e error) {
	var id ID
	if id, e = NewID(); check(e) {
		return
	}
	n = &Node{
		ID:       id,
		IP:       ip,
		Sessions: session.Sessions{},
	}
	return
}

type Nodes []*Node

// NewNodes creates an empty Nodes
func NewNodes() (n Nodes) { return Nodes{} }

func (n Nodes) Len() int {
	return len(n)
}

// Add a Node to a Nodes.
func (n Nodes) Add(nn *Node) Nodes {
	return append(n, nn)
}

// FindByID searches for a Node by ID.
func (n Nodes) FindByID(id ID) (no *Node) {
	for _, nn := range n {
		if nn.ID == id {
			no = nn
			break
		}
	}
	return
}

// FindByIP searches for a Node by net.IP.
func (n Nodes) FindByIP(id net.IP) (no *Node) {
	for _, nn := range n {
		if nn.IP.String() == id.String() {
			no = nn
			break
		}
	}
	return
}

// DeleteByID deletes a node identified by an ID.
func (n Nodes) DeleteByID(id ID) (nn Nodes, e error) {
	e = fmt.Errorf("id %x not found", id)
	for i := range n {
		if n[i].ID == id {
			n = append(n[:i], n[i+1:]...)
			e = nil
			break
		}
	}
	return
}

// DeleteByIP deletes a node identified by a net.IP.
func (n Nodes) DeleteByIP(ip net.IP) (nn Nodes, e error) {
	e = fmt.Errorf("node with ip %v not found", ip)
	for i := range n {
		if n[i].IP.String() == ip.String() {
			n = append(n[:i], n[i+1:]...)
			e = nil
			break
		}
	}
	return
}
