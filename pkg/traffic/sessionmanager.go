package traffic

import (
	"sync"

	"git-indra.lan/indra-labs/lnd/lnd/lnwire"

	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
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

type SessionManager struct {
	nodes           []*Node
	pendingPayments PendingPayments
	Sessions
	SessionCache
	sync.Mutex
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		SessionCache: make(SessionCache),
	}
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

func (sm *SessionManager) GetNodeCircuit(id nonce.ID) (sce *Circuit,
	exists bool) {

	sm.Lock()
	defer sm.Unlock()
	sce, exists = sm.SessionCache[id]
	return
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
	// Hop 5, the return session( s) are not added to the SessionCache as they
	// are not billable and are only related to the node of the Engine.
	if s.Hop < 5 {
		sm.SessionCache.Add(s)
	}
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
			// Delete from Session cache.
			sm.SessionCache[sm.Sessions[i].Node.ID][sm.Sessions[i].Hop] = nil
			// Delete from Sessions.
			sm.Sessions = append(sm.Sessions[:i], sm.Sessions[i+1:]...)
		}
	}
}

// IterateSessions calls a function for each entry in the Sessions slice.
//
// Do not call SessionManager methods within this function.
func (sm *SessionManager) IterateSessions(fn func(s *Session) bool) {
	sm.Lock()
	defer sm.Unlock()
	for i := range sm.Sessions {
		if fn(sm.Sessions[i]) {
			break
		}
	}
}

// IterateSessionCache calls a function for each entry in the SessionCache
// that provides also access to the related node.
//
// Do not call SessionManager methods within this function.
func (sm *SessionManager) IterateSessionCache(fn func(n *Node,
	c *Circuit) bool) {

	sm.Lock()
	defer sm.Unlock()
out:
	for i := range sm.SessionCache {
		for j := range sm.nodes {
			if sm.nodes[j].ID == i {
				if fn(sm.nodes[j], sm.SessionCache[i]) {
					break out
				}
				break
			}
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

func (sm *SessionManager) DeleteNodeAndSessions(id nonce.ID) {
	sm.Lock()
	defer sm.Unlock()
	var exists bool
	// If the node exists its ID is in the SessionCache.
	if _, exists = sm.SessionCache[id]; !exists {
		return
	}
	delete(sm.SessionCache, id)
	// Delete from the nodes list.
	for i := range sm.nodes {
		if sm.nodes[i].ID == id {
			sm.nodes = append(sm.nodes[:i], sm.nodes[i+1:]...)
			break
		}
	}
	var found []int
	// Locate all the sessions with the node in them.
	for i := range sm.Sessions {
		if sm.Sessions[i].Node.ID == id {
			found = append(found, i)
		}
	}
	// Create a new Sessions slice and add the ones not in the found list.
	temp := make(Sessions, 0, len(sm.Sessions)-len(found))
	for i := range sm.Sessions {
		for j := range found {
			if i != found[j] {
				temp = append(temp, sm.Sessions[i])
				break
			}
		}
	}
	// Place the new Sessions slice in place of the old.
	sm.Sessions = temp
}
