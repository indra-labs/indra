package client

import (
	"github.com/indra-labs/indra/pkg/onion/layers/response"
	"github.com/indra-labs/indra/pkg/types"
	"github.com/indra-labs/indra/pkg/util/slice"
	"github.com/indra-labs/lnd/lnd/lnwire"
)

// response is a payload from an exit message.
func (cl *Client) response(on *response.Layer, b slice.Bytes,
	cur *slice.Cursor, prev types.Onion) {

	pending := cl.PendingResponses.Find(on.Hash)
	first := true
	var rr lnwire.MilliSatoshi
	if pending != nil {
		for i := range pending.Billable {
			if first {
				first = false
				s := cl.FindSession(pending.Billable[i])
				for i := range s.Services {
					if s.Services[i].Port == pending.Port {
						rr = s.Services[i].RelayRate
					}
				}
				if s != nil {
					log.D.Ln(cl.AddrPort.String(), "exit send", i)
					cl.DecSession(s.ID, rr*lnwire.
						MilliSatoshi(len(b)/2)/1024/1024)
				}
				continue
			}
			s := cl.FindSession(pending.Billable[i])
			if s != nil {
				log.D.Ln(cl.AddrPort.String(), "reverse")
				cl.DecSession(s.ID, s.RelayRate*lnwire.
					MilliSatoshi(len(b))/1024/1024)
			}
		}
		pending.Callback(on.Bytes)
		cl.PendingResponses.Delete(on.Hash)
	}
}
