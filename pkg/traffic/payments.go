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

type PaymentChan chan *payment.Payment

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
