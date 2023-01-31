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

type SessionCache map[nonce.ID]SessionCacheEntry

type SessionCacheEntry struct {
	Hops [5]*Session
}

type SessionManager struct {
	nodes           []*Node
	pendingPayments PendingPayments
	Sessions
	PaymentChan
	SessionCache
	sync.Mutex
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		PaymentChan:  make(PaymentChan),
		SessionCache: make(SessionCache),
	}
}

func (sm *SessionManager) UpdateSessionCache() {

}

func (sm *SessionManager) IncSession(id nonce.ID, sats lnwire.MilliSatoshi,
	sender bool, typ string) {
	sess := sm.FindSession(id)
	if sess != nil {
		sm.Lock()
		defer sm.Unlock()
		sess.IncSats(sats, sender, typ)
	}
}
func (sm *SessionManager) DecSession(id nonce.ID, sats lnwire.MilliSatoshi,
	sender bool, typ string) bool {
	sess := sm.FindSession(id)
	if sess != nil {
		sm.Lock()
		defer sm.Unlock()
		return sess.DecSats(sats, sender, typ)
	}
	return false
}

func (sm *SessionManager) AddSession(s *Session) {
	sm.Lock()
	defer sm.Unlock()
	// check for dupes
	for i := range sm.Sessions {
		if sm.Sessions[i].ID == s.ID {
			log.D.Ln("refusing to add duplicate session ID")
			return
		}
	}
	sm.Sessions = append(sm.Sessions, s)
}
func (sm *SessionManager) FindSession(id nonce.ID) *Session {
	sm.Lock()
	defer sm.Unlock()
	for i := range sm.Sessions {
		if sm.Sessions[i].ID == id {
			return sm.Sessions[i]
		}
	}
	return nil
}
func (sm *SessionManager) FindSessionByHeader(prvKey *prv.Key) *Session {
	sm.Lock()
	defer sm.Unlock()
	for i := range sm.Sessions {
		if sm.Sessions[i].HeaderPrv.Key.Equals(&prvKey.Key) {
			return sm.Sessions[i]
		}
	}
	return nil
}
func (sm *SessionManager) FindSessionByHeaderPub(pubKey *pub.Key) *Session {
	sm.Lock()
	defer sm.Unlock()
	for i := range sm.Sessions {
		if sm.Sessions[i].HeaderPub.Equals(pubKey) {
			return sm.Sessions[i]
		}
	}
	return nil
}
func (sm *SessionManager) FindSessionPreimage(pr sha256.Hash) *Session {
	sm.Lock()
	defer sm.Unlock()
	for i := range sm.Sessions {
		if sm.Sessions[i].Preimage == pr {
			return sm.Sessions[i]
		}
	}
	return nil
}
func (sm *SessionManager) GetSessionsAtHop(hop byte) (s Sessions) {
	sm.Lock()
	defer sm.Unlock()
	for i := range sm.Sessions {
		if sm.Sessions[i].Hop == hop {
			s = append(s, sm.Sessions[i])
		}
	}
	return
}
func (sm *SessionManager) DeleteSession(id nonce.ID) {
	sm.Lock()
	defer sm.Unlock()
	for i := range sm.Sessions {
		if sm.Sessions[i].ID == id {
			sm.Sessions = append(sm.Sessions[:i], sm.Sessions[i+1:]...)
		}
	}

}
func (sm *SessionManager) IterateSessions(fn func(s *Session) bool) {
	sm.Lock()
	defer sm.Unlock()
	for i := range sm.Sessions {
		if fn(sm.Sessions[i]) {
			break
		}
	}
}

func (sm *SessionManager) GetSessionByIndex(i int) (s *Session) {
	sm.Lock()
	defer sm.Unlock()
	if len(sm.Sessions) > i {
		s = sm.Sessions[i]
	}
	return
}
