package relay

import (
	"git-indra.lan/indra-labs/lnd/lnd/lnwire"
	
	"git-indra.lan/indra-labs/indra/pkg/onion/balance"
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
				switch {
				case i < 2:
					in := s.RelayRate * lnwire.MilliSatoshi(
						pending.SentSize) / 1024 / 1024
					eng.DecSession(s.ID, in, true, "reverse")
				case i == 2:
					in := s.RelayRate * lnwire.MilliSatoshi(
						pending.SentSize/2) / 1024 / 1024
					out := s.RelayRate * lnwire.MilliSatoshi(
						len(b)/2) / 1024 / 1024
					eng.DecSession(s.ID, in+out, true, "getbalance")
				case i > 2:
					out := s.RelayRate * lnwire.MilliSatoshi(
						len(b)) / 1024 / 1024
					eng.DecSession(s.ID, out, true, "reverse")
				}
			}
		}
		var se *Session
		eng.IterateSessions(func(s *Session) bool {
			if s.ID == on.ID {
				log.D.F("%s received balance %s for session %s %s was %s",
					local,
					on.MilliSatoshi, on.ID, on.ConfID, s.Remaining)
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
