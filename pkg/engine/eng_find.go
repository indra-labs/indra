package engine

import (
	"git-indra.lan/indra-labs/indra/pkg/crypto"
)

// FindCloaked searches the client identity key and the sessions for a match. It
// returns the session as well, though not all users of this function will need
// this.
func (sm *SessionManager) FindCloaked(clk crypto.PubKey) (hdr *crypto.Prv,
	pld *crypto.Prv, sess *SessionData, identity bool) {
	
	var b crypto.Blinder
	copy(b[:], clk[:crypto.BlindLen])
	hash := crypto.Cloak(b, sm.GetLocalNodeIdentityBytes())
	if hash == clk {
		log.T.F("encrypted to identity key")
		hdr = sm.GetLocalNodeIdentityPrv()
		// there is no payload key for the node, only in 
		identity = true
		return
	}
	sm.IterateSessions(func(s *SessionData) (stop bool) {
		hash = crypto.Cloak(b, s.Header.Bytes)
		if hash == clk {
			hdr = s.Header.Prv
			pld = s.Payload.Prv
			sess = s
			return true
		}
		return
	})
	return
}
