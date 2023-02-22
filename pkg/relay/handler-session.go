package relay

import (
	"git-indra.lan/indra-labs/indra/pkg/onion/session"
	"git-indra.lan/indra-labs/indra/pkg/types"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

func (eng *Engine) session(on *session.Layer, b slice.Bytes,
	c *slice.Cursor, prev types.Onion) {
	
	log.D.Ln(prev == nil)
	log.T.F("incoming session %s", on.ID)
	pi := eng.FindPendingPreimage(on.PreimageHash())
	if pi != nil {
		// We need to delete this first in case somehow two such messages arrive
		// at the same time, and we end up with duplicate sessions.
		eng.DeletePendingPayment(pi.Preimage)
		log.D.F("Adding session %s to %s", pi.ID, eng.GetLocalNodeAddress())
		eng.AddSession(NewSession(pi.ID,
			eng.GetLocalNode(), pi.Amount, on.Header, on.Payload, on.Hop))
		eng.handleMessage(BudgeUp(b, *c), on)
	} else {
		log.E.Ln("dropping session message without payment")
	}
}
