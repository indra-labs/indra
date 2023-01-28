package indra

import (
	"github.com/indra-labs/indra/pkg/crypto/key/cloak"
	"github.com/indra-labs/indra/pkg/crypto/key/prv"
	"github.com/indra-labs/indra/pkg/traffic"
)

// FindCloaked searches the client identity key and the sessions for a match. It
// returns the session as well, though not all users of this function will need
// this.
func (en *Engine) FindCloaked(clk cloak.PubKey) (hdr *prv.Key,
	pld *prv.Key, sess *traffic.Session, identity bool) {

	var b cloak.Blinder
	copy(b[:], clk[:cloak.BlindLen])
	hash := cloak.Cloak(b, en.Node.IdentityBytes)
	if hash == clk {
		log.T.F("encrypted to identity key")
		hdr = en.Node.IdentityPrv
		// there is no payload key for the node, only in sessions.
		identity = true
		return
	}
	var i int
	en.Node.IterateSessions(func(s *traffic.Session) (stop bool) {
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
