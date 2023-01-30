package relay

import (
	"git-indra.lan/indra-labs/lnd/lnd/lnwire"

	"git-indra.lan/indra-labs/indra/pkg/onion/layers/response"
	"git-indra.lan/indra-labs/indra/pkg/types"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

// response is a payload from an exit message.
func (en *Engine) response(on *response.Layer, b slice.Bytes,
	cur *slice.Cursor, prev types.Onion) {

	pending := en.Pending.Find(on.ID)
	log.T.F("searching for pending ID %x", on.ID)
	if pending != nil {
		for i := range pending.Billable {
			s := en.FindSession(pending.Billable[i])
			if s != nil {
				typ := "response"
				relayRate := s.Peer.RelayRate
				dataSize := len(b)
				switch i {
				case 0, 1:
					dataSize = pending.SentSize
				case 2:
					for j := range s.Peer.Services {
						if s.Peer.Services[j].Port == on.Port {
							relayRate = s.Peer.Services[j].RelayRate / 2
							typ = "exit"
						}
					}
				}
				en.DecSession(s.ID, relayRate*lnwire.
					MilliSatoshi(dataSize)/1024/1024, true, typ)
			}
		}
		en.Pending.Delete(on.ID, on.Bytes)
	}
}
