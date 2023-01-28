package indra

import (
	"github.com/indra-labs/lnd/lnd/lnwire"

	"github.com/indra-labs/indra/pkg/onion/layers/response"
	"github.com/indra-labs/indra/pkg/types"
	"github.com/indra-labs/indra/pkg/util/slice"
)

// response is a payload from an exit message.
func (en *Engine) response(on *response.Layer, b slice.Bytes,
	cur *slice.Cursor, prev types.Onion) {

	pending := en.Pending.Find(on.Hash)
	first := true
	var rr lnwire.MilliSatoshi
	if pending != nil {
		for i := range pending.Billable {
			if first {
				first = false
				s := en.FindSession(pending.Billable[i])
				for i := range s.Services {
					if s.Services[i].Port == pending.Port {
						rr = s.Services[i].RelayRate
					}
				}
				if s != nil {
					log.D.Ln(en.AddrPort.String(), "exit send", i)
					en.DecSession(s.ID, rr*lnwire.
						MilliSatoshi(len(b)/2)/1024/1024)
				}
				continue
			}
			s := en.FindSession(pending.Billable[i])
			if s != nil {
				log.D.Ln(en.AddrPort.String(), "reverse")
				en.DecSession(s.ID, s.RelayRate*lnwire.
					MilliSatoshi(len(b))/1024/1024)
			}
		}
		pending.Callback(on.Bytes)
		en.Pending.Delete(on.Hash)
	}
}
