package traffic

import (
	"sync"

	"github.com/indra-labs/indra"
	"github.com/indra-labs/indra/pkg/crypto/key/prv"
	"github.com/indra-labs/indra/pkg/crypto/key/pub"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/crypto/sha256"
	"github.com/indra-labs/indra/pkg/identity"
	"github.com/indra-labs/indra/pkg/payment"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/lnd/lnd/lnwire"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

// A Session keeps track of a connection session. It specifically maintains the
// account of available bandwidth allocation before it needs to be recharged
// with new credit, and the current state of the encryption.
type Session struct {
	nonce.ID
	*identity.Peer
	Remaining                 lnwire.MilliSatoshi
	HeaderPrv, PayloadPrv     *prv.Key
	HeaderPub, PayloadPub     *pub.Key
	HeaderBytes, PayloadBytes pub.Bytes
	Preimage                  sha256.Hash
	Hop                       byte
}

// NewSession creates a new Session.
//
// Purchasing a session the seller returns a token, based on a requested data
// allocation.
func NewSession(
	id nonce.ID,
	node *identity.Peer,
	rem lnwire.MilliSatoshi,
	hdrPrv *prv.Key,
	pldPrv *prv.Key,
	hop byte,
) (s *Session) {

	var e error
	if hdrPrv == nil || pldPrv == nil {
		if hdrPrv, e = prv.GenerateKey(); check(e) {
		}
		if pldPrv, e = prv.GenerateKey(); check(e) {
		}
	}
	hdrPub := pub.Derive(hdrPrv)
	pldPub := pub.Derive(pldPrv)
	h, p := hdrPrv.ToBytes(), pldPrv.ToBytes()
	s = &Session{
		ID:           id,
		Peer:         node,
		Remaining:    rem,
		HeaderPub:    hdrPub,
		HeaderBytes:  hdrPub.ToBytes(),
		PayloadPub:   pldPub,
		PayloadBytes: pldPub.ToBytes(),
		HeaderPrv:    hdrPrv,
		PayloadPrv:   pldPrv,
		Preimage:     sha256.Single(append(h[:], p[:]...)),
		Hop:          hop,
	}
	return
}

// AddBytes adds to the Remaining counter, used when new data allowance has been
// purchased.
func (s *Session) AddBytes(b lnwire.MilliSatoshi) { s.Remaining += b }

// SubtractBytes reduces the amount Remaining, if the requested amount would put
// the total below zero it returns false, signalling that new data allowance
// needs to be purchased before any further messages can be sent.
func (s *Session) SubtractBytes(b lnwire.MilliSatoshi) bool {
	if s.Remaining < b {
		return false
	}
	s.Remaining -= b
	return true
}

type Circuit [5]*Session

type Sessions []*Session

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

type Payments struct {
	pendingPayments PendingPayments
	sessions        Sessions
	PaymentChan
	sync.Mutex
}

func NewPayments() *Payments {
	return &Payments{PaymentChan: make(PaymentChan)}
}

func (n *Payments) AddSession(s *Session) {
	n.Lock()
	defer n.Unlock()
	// check for dupes
	for i := range n.sessions {
		if n.sessions[i].ID == s.ID {
			log.D.Ln("refusing to add duplicate session ID")
			return
		}
	}
	n.sessions = append(n.sessions, s)
}
func (n *Payments) FindSession(id nonce.ID) *Session {
	n.Lock()
	defer n.Unlock()
	for i := range n.sessions {
		if n.sessions[i].ID == id {
			return n.sessions[i]
		}
	}
	return nil
}
func (n *Payments) GetSessionsAtHop(hop byte) (s Sessions) {
	n.Lock()
	defer n.Unlock()
	for i := range n.sessions {
		if n.sessions[i].Hop == hop {
			s = append(s, n.sessions[i])
		}
	}
	return
}
func (n *Payments) DeleteSession(id nonce.ID) {
	n.Lock()
	defer n.Unlock()
	for i := range n.sessions {
		if n.sessions[i].ID == id {
			n.sessions = append(n.sessions[:i], n.sessions[i+1:]...)
		}
	}

}
func (n *Payments) IterateSessions(fn func(s *Session) bool) {
	n.Lock()
	defer n.Unlock()
	for i := range n.sessions {
		if fn(n.sessions[i]) {
			break
		}
	}
}

func (n *Payments) GetSessionByIndex(i int) (s *Session) {
	n.Lock()
	defer n.Unlock()
	if len(n.sessions) > i {
		s = n.sessions[i]
	}
	return
}

type PaymentChan chan *payment.Payment

type PendingPayments []*payment.Payment

func (p PendingPayments) Add(np *payment.Payment) (pp PendingPayments) {
	return append(p, np)
}

func (p PendingPayments) Delete(preimage sha256.Hash) (pp PendingPayments) {
	pp = p
	for i := range p {
		if p[i].Preimage == preimage {
			if i == len(p)-1 {
				pp = p[:i]
			} else {
				pp = append(p[:i], p[i+1:]...)
			}
		}
	}
	return
}

func (p PendingPayments) Find(id nonce.ID) (pp *payment.Payment) {
	for i := range p {
		if p[i].ID == id {
			return p[i]
		}
	}
	return
}

func (p PendingPayments) FindPreimage(pi sha256.Hash) (pp *payment.Payment) {
	for i := range p {
		if p[i].Preimage == pi {
			return p[i]
		}
	}
	return
}

// PendingPayment accessors. For the same reason as the sessions, pending
// payments need to be accessed only with the node's mutex locked.

func (n *Payments) AddPendingPayment(
	np *payment.Payment) {

	n.Lock()
	defer n.Unlock()
	n.pendingPayments = n.pendingPayments.Add(np)
}
func (n *Payments) DeletePendingPayment(
	preimage sha256.Hash) {

	n.Lock()
	defer n.Unlock()
	n.pendingPayments = n.pendingPayments.Delete(preimage)
}
func (n *Payments) FindPendingPayment(
	id nonce.ID) (pp *payment.Payment) {

	n.Lock()
	defer n.Unlock()
	return n.pendingPayments.Find(id)
}
func (n *Payments) FindPendingPreimage(
	pi sha256.Hash) (pp *payment.Payment) {

	log.T.F("searching preimage %x", pi)
	n.Lock()
	defer n.Unlock()
	return n.pendingPayments.FindPreimage(pi)
}
