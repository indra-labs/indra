package relay

import (
	"git-indra.lan/indra-labs/lnd/lnd/lnwire"

	"git-indra.lan/indra-labs/indra/pkg/onion/layers/response"
	"git-indra.lan/indra-labs/indra/pkg/types"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

// response is a payload from an exit message.
func (eng *Engine) response(on *response.Layer, b slice.Bytes,
	cur *slice.Cursor, prev types.Onion) {

	pending := eng.PendingResponses.Find(on.ID)
	log.T.F("searching for pending ID %x", on.ID)
	if pending != nil {
		for i := range pending.Billable {
			s := eng.FindSession(pending.Billable[i])
			if s != nil {
				typ := "response"
				relayRate := s.RelayRate
				dataSize := len(b)
				switch i {
				case 0, 1:
					dataSize = pending.SentSize
				case 2:
					for j := range s.Services {
						if s.Services[j].Port == on.Port {
							relayRate = s.Services[j].RelayRate / 2
							typ = "exit"
						}
					}
				}
				eng.DecSession(s.ID, relayRate*lnwire.
					MilliSatoshi(dataSize)/1024/1024, true, typ)
			}
		}
		eng.PendingResponses.Delete(on.ID, on.Bytes)
	}
}
