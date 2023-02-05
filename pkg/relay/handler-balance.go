package relay

import (
	"git-indra.lan/indra-labs/lnd/lnd/lnwire"
	
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/balance"
	"git-indra.lan/indra-labs/indra/pkg/traffic"
	"git-indra.lan/indra-labs/indra/pkg/types"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

func (eng *Engine) balance(on *balance.Layer,
	b slice.Bytes, c *slice.Cursor, prev types.Onion) {
	
	local := eng.GetLocalNodeAddress()
	pending := eng.PendingResponses.Find(on.ID)
	if pending != nil {
		for i := range pending.Billable {
			s := eng.FindSession(pending.Billable[i])
			if s != nil {
				if i == 0 {
					eng.DecSession(s.ID,
						s.RelayRate*lnwire.MilliSatoshi(len(b)/2)/1024/1024,
						true, "balance1")
				} else {
					eng.DecSession(s.ID,
						s.RelayRate*lnwire.MilliSatoshi(len(b))/1024/1024,
						true, "balance2")
				}
			}
		}
		var se *traffic.Session
		eng.IterateSessions(func(s *traffic.Session) bool {
			if s.ID == on.ID {
				log.D.F("%s received balance %s for session %s was %s", local,
					on.MilliSatoshi, on.ID, s.Remaining)
				se = s
				return true
			}
			return false
		})
		eng.PendingResponses.Delete(pending.ID, nil)
		if se != nil {
			log.D.F("got %v, expected %v", se.Remaining, on.MilliSatoshi)
			se.Remaining = on.MilliSatoshi
		}
	}
}
