package relay

import (
	"git-indra.lan/indra-labs/lnd/lnd/lnwire"

	"git-indra.lan/indra-labs/indra/pkg/onion/layers/balance"
	"git-indra.lan/indra-labs/indra/pkg/traffic"
	"git-indra.lan/indra-labs/indra/pkg/types"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

func (en *Engine) balance(on *balance.Layer,
	b slice.Bytes, c *slice.Cursor, prev types.Onion) {

	en.IterateSessions(func(s *traffic.Session) bool {
		if s.ID == on.ID {
			log.D.F("received balance %x for session %x",
				on.MilliSatoshi, on.ID)
			s.Remaining = on.MilliSatoshi
			return true
		}
		return false
	})
	pending := en.Pending.Find(on.ID)
	if pending != nil {
		for i := range pending.Billable {
			s := en.FindSession(pending.Billable[i])
			if s != nil {
				if i == 0 {
					en.DecSession(s.ID,
						s.RelayRate*lnwire.MilliSatoshi(len(b)/2)/1024/1024,
						true, "balance")
				} else {
					en.DecSession(s.ID,
						s.RelayRate*lnwire.MilliSatoshi(len(b))/1024/1024,
						true, "balance")
				}
			}
		}
		en.Pending.Delete(pending.ID, nil)
	}
}
