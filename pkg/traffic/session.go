package traffic

import (
	"sync"

	"git-indra.lan/indra-labs/lnd/lnd/lnwire"

	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/payment"
)

// A Session keeps track of a connection session. It specifically maintains the
// account of available bandwidth allocation before it needs to be recharged
// with new credit, and the current state of the encryption.
type Session struct {
	ID nonce.ID
	*Node
	Remaining                 lnwire.MilliSatoshi
	HeaderPrv, PayloadPrv     *prv.Key
	HeaderPub, PayloadPub     *pub.Key
	HeaderBytes, PayloadBytes pub.Bytes
	Preimage                  sha256.Hash
	Hop                       byte
}

// NewSession creates a new Session.
func NewSession(
	id nonce.ID,
	node *Node,
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
		Node:         node,
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

// IncSats adds to the Remaining counter, used when new data allowance has been
// purchased.
func (s *Session) IncSats(sats lnwire.MilliSatoshi,
	sender bool, typ string) {
	who := "relay"
	if sender {
		who = "client"
	}
	log.D.F("%s session %d %x current %v incrementing by %v", who, typ, s.ID,
		s.Remaining, sats)
	s.Remaining += sats
}

// DecSats reduces the amount Remaining, if the requested amount would put
// the total below zero it returns false, signalling that new data allowance
// needs to be purchased before any further messages can be sent.
func (s *Session) DecSats(sats lnwire.MilliSatoshi,
	sender bool, typ string) bool {
	if s.Remaining < sats {
		return false
	}
	who := "relay"
	if sender {
		who = "client"
	}
	log.D.F("%s session %s %x current %v decrementing by %v", who, typ, s.ID,
		s.Remaining, sats)
	s.Remaining -= sats
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
	Sessions
	PaymentChan
	sync.Mutex
}

func NewPayments() *Payments {
	return &Payments{PaymentChan: make(PaymentChan)}
}

func (pm *Payments) IncSession(id nonce.ID, sats lnwire.MilliSatoshi,
	sender bool, typ string) {
	sess := pm.FindSession(id)
	if sess != nil {
		pm.Lock()
		defer pm.Unlock()
		sess.IncSats(sats, sender, typ)
	}
}
func (pm *Payments) DecSession(id nonce.ID, sats lnwire.MilliSatoshi,
	sender bool, typ string) bool {
	sess := pm.FindSession(id)
	if sess != nil {
		pm.Lock()
		defer pm.Unlock()
		return sess.DecSats(sats, sender, typ)
	}
	return false
}

func (pm *Payments) AddSession(s *Session) {
	pm.Lock()
	defer pm.Unlock()
	// check for dupes
	for i := range pm.Sessions {
		if pm.Sessions[i].ID == s.ID {
			log.D.Ln("refusing to add duplicate session ID")
			return
		}
	}
	pm.Sessions = append(pm.Sessions, s)
}
func (pm *Payments) FindSession(id nonce.ID) *Session {
	pm.Lock()
	defer pm.Unlock()
	for i := range pm.Sessions {
		if pm.Sessions[i].ID == id {
			return pm.Sessions[i]
		}
	}
	return nil
}
func (pm *Payments) FindSessionByHeader(prvKey *prv.Key) *Session {
	pm.Lock()
	defer pm.Unlock()
	for i := range pm.Sessions {
		if pm.Sessions[i].HeaderPrv.Key.Equals(&prvKey.Key) {
			return pm.Sessions[i]
		}
	}
	return nil
}
func (pm *Payments) FindSessionByHeaderPub(pubKey *pub.Key) *Session {
	pm.Lock()
	defer pm.Unlock()
	for i := range pm.Sessions {
		if pm.Sessions[i].HeaderPub.Equals(pubKey) {
			return pm.Sessions[i]
		}
	}
	return nil
}
func (pm *Payments) FindSessionPreimage(pr sha256.Hash) *Session {
	pm.Lock()
	defer pm.Unlock()
	for i := range pm.Sessions {
		if pm.Sessions[i].Preimage == pr {
			return pm.Sessions[i]
		}
	}
	return nil
}
func (pm *Payments) GetSessionsAtHop(hop byte) (s Sessions) {
	pm.Lock()
	defer pm.Unlock()
	for i := range pm.Sessions {
		if pm.Sessions[i].Hop == hop {
			s = append(s, pm.Sessions[i])
		}
	}
	return
}
func (pm *Payments) DeleteSession(id nonce.ID) {
	pm.Lock()
	defer pm.Unlock()
	for i := range pm.Sessions {
		if pm.Sessions[i].ID == id {
			pm.Sessions = append(pm.Sessions[:i], pm.Sessions[i+1:]...)
		}
	}

}
func (pm *Payments) IterateSessions(fn func(s *Session) bool) {
	pm.Lock()
	defer pm.Unlock()
	for i := range pm.Sessions {
		if fn(pm.Sessions[i]) {
			break
		}
	}
}

func (pm *Payments) GetSessionByIndex(i int) (s *Session) {
	pm.Lock()
	defer pm.Unlock()
	if len(pm.Sessions) > i {
		s = pm.Sessions[i]
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

func (pm *Payments) AddPendingPayment(
	np *payment.Payment) {

	pm.Lock()
	defer pm.Unlock()
	pm.pendingPayments = pm.pendingPayments.Add(np)
}
func (pm *Payments) DeletePendingPayment(
	preimage sha256.Hash) {

	pm.Lock()
	defer pm.Unlock()
	pm.pendingPayments = pm.pendingPayments.Delete(preimage)
}
func (pm *Payments) FindPendingPayment(
	id nonce.ID) (pp *payment.Payment) {

	pm.Lock()
	defer pm.Unlock()
	return pm.pendingPayments.Find(id)
}
func (pm *Payments) FindPendingPreimage(
	pi sha256.Hash) (pp *payment.Payment) {

	log.T.F("searching preimage %x", pi)
	pm.Lock()
	defer pm.Unlock()
	return pm.pendingPayments.FindPreimage(pi)
}
