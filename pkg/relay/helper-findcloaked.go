package relay

import (
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/cloak"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/traffic"
)

// FindCloaked searches the client identity key and the sessions for a match. It
// returns the session as well, though not all users of this function will need
// this.
func (eng *Engine) FindCloaked(clk cloak.PubKey) (hdr *prv.Key,
	pld *prv.Key, sess *traffic.Session, identity bool) {

	var b cloak.Blinder
	copy(b[:], clk[:cloak.BlindLen])
	hash := cloak.Cloak(b, eng.GetLocalNode().IdentityBytes)
	if hash == clk {
		log.T.F("encrypted to identity key")
		hdr = eng.GetLocalNode().IdentityPrv
		// there is no payload key for the node, only in sessions.
		identity = true
		return
	}
	var i int
	eng.IterateSessions(func(s *traffic.Session) (stop bool) {
		hash = cloak.Cloak(b, s.HeaderBytes)
		if hash == clk {
			log.T.F("found cloaked key in session %d", i)
			hdr = s.HeaderPrv
			pld = s.PayloadPrv
			sess = s
			return true
		}
		i++
		return
	})
	return
}
