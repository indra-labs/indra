package traffic

import (
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
)

// A SessionCache stores each of the 5 hops
type SessionCache map[nonce.ID]*Circuit

func (sm *SessionManager) UpdateSessionCache() {
	sm.Lock()
	defer sm.Unlock()
	// First we create SessionCache entries for all existing nodes.
	for i := range sm.nodes {
		_, exists := sm.SessionCache[sm.nodes[i].ID]
		if !exists {
			sm.SessionCache[sm.nodes[i].ID] = &Circuit{}
		}
	}
	// Place all sessions in their slots respective to their node.
	for _, v := range sm.Sessions {
		sm.SessionCache[v.Node.ID][v.Hop] = v
	}
}

func (sc SessionCache) Add(s *Session) SessionCache {
	log.D.F("adding session %s %s %d", s.ID, s.AddrPort.String(), s.Hop)
	var sce *Circuit
	var exists bool
	if sce, exists = sc[s.Node.ID]; !exists {
		sce = &Circuit{}
		sce[s.Hop] = s
		sc[s.Node.ID] = sce
		return sc
	}
	sc[s.Node.ID][s.Hop] = s
	log.D.S(sc[s.Node.ID])
	return sc
}
