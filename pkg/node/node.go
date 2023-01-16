// Package node maintains information about peers on the network and associated
// connection sessions.
package node

import (
	"fmt"
	"net/netip"
	sync "sync"
	"time"

	"github.com/indra-labs/indra"
	"github.com/indra-labs/indra/pkg/ifc"
	"github.com/indra-labs/indra/pkg/key/prv"
	"github.com/indra-labs/indra/pkg/key/pub"
	log2 "github.com/indra-labs/indra/pkg/log"
	"github.com/indra-labs/indra/pkg/nonce"
	"github.com/indra-labs/indra/pkg/sha256"
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
	Addr            string
	AddrPort        *netip.AddrPort
	IdentityPub     *pub.Key
	IdentityBytes   pub.Bytes
	IdentityPrv     *prv.Key
	PingCount       int
	LastSeen        time.Time
	pendingPayments PendingPayments
	sessions        Sessions
	Services
	PaymentChan
	ifc.Transport
}

// New creates a new Node. netip.AddrPort is optional if the counterparty is not
// in direct connection. Also, the idPrv node private key can be nil, as only
// the node embedded in a client and not the peer node list has one available.
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
		PaymentChan:   make(PaymentChan),
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
			log.T.Ln("receivefrom")
			b = n.Services[i].Receive()
			return
		}
	}
	return
}

// Session management functions
//
// In order to enable scaling client processing the session data will be
// accessed by multiple goroutines, and thus we use the node's mutex to prevent
// race conditions on this data. This is the only mutable data. A relay's
// identity is its keys so a different key is a different relay, even if it is
// on the same IP address. Because we use netip.AddrPort as network addresses
// there can be more than one relay at an IP address, though hop selection will
// consider the IP address as the unique identifier and not select more than one
// relay on the same IP address. (todo:)

func (n *Node) AddSession(s *Session) {
	n.Lock()
	defer n.Unlock()
	n.sessions = append(n.sessions, s)
}
func (n *Node) FindSession(id nonce.ID) *Session {
	n.Lock()
	defer n.Unlock()
	for i := range n.sessions {
		if n.sessions[i].ID == id {
			return n.sessions[i]
		}
	}
	return nil
}
func (n *Node) GetSessionsAtHop(hop byte) (s Sessions) {
	n.Lock()
	defer n.Unlock()
	for i := range n.sessions {
		if n.sessions[i].Hop == hop {
			s = append(s, n.sessions[i])
		}
	}
	return
}
func (n *Node) DeleteSession(id nonce.ID) {
	n.Lock()
	defer n.Unlock()
	for i := range n.sessions {
		if n.sessions[i].ID == id {
			n.sessions = append(n.sessions[:i], n.sessions[i+1:]...)
		}
	}

}
func (n *Node) IterateSessions(fn func(s *Session) bool) {
	n.Lock()
	defer n.Unlock()
	for i := range n.sessions {
		if fn(n.sessions[i]) {
			break
		}
	}
}

func (n *Node) GetSessionByIndex(i int) (s *Session) {
	n.Lock()
	defer n.Unlock()
	if len(n.sessions) > i {
		s = n.sessions[i]
	}
	return
}

// PendingPayment accessors. For the same reason as the sessions, pending
// payments need to be accessed only with the node's mutex locked.

func (n *Node) AddPendingPayment(
	np *Payment) {

	n.Lock()
	defer n.Unlock()
	n.pendingPayments = n.pendingPayments.Add(np)
}
func (n *Node) DeletePendingPayment(
	preimage sha256.Hash) {

	n.Lock()
	defer n.Unlock()
	n.pendingPayments = n.pendingPayments.Delete(preimage)
}
func (n *Node) FindPendingPayment(
	id nonce.ID) (pp *Payment) {

	n.Lock()
	defer n.Unlock()
	return n.pendingPayments.Find(id)
}
func (n *Node) FindPendingPreimage(
	pi sha256.Hash) (pp *Payment) {

	log.T.F("searching preimage %x", pi)
	n.Lock()
	defer n.Unlock()
	return n.pendingPayments.FindPreimage(pi)
}
