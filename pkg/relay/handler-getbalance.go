package relay

import (
	"git-indra.lan/indra-labs/lnd/lnd/lnwire"

	"git-indra.lan/indra-labs/indra/pkg/onion/balance"
	"git-indra.lan/indra-labs/indra/pkg/onion/crypt"
	"git-indra.lan/indra-labs/indra/pkg/onion/getbalance"
	"git-indra.lan/indra-labs/indra/pkg/types"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

func (eng *Engine) getBalance(on *getbalance.Layer,
	b slice.Bytes, c *slice.Cursor, prev types.Onion) {

	log.T.S(on)
	var found bool
	var bal *balance.Layer
	eng.IterateSessions(func(s *Session) bool {
		if s.ID == on.ID {
			bal = &balance.Layer{
				ID:           on.ID,
				ConfID:       on.ConfID,
				MilliSatoshi: s.Remaining,
			}
			found = true
			return true
		}
		return false
	})
	if !found {
		log.E.Ln("session not found", on.ID)
		log.D.S(eng.Sessions)
		return
	}
	header := b[*c:c.Inc(crypt.ReverseHeaderLen)]
	rb := FormatReply(header,
		Encode(bal), on.Ciphers, on.Nonces)
	rb = append(rb, slice.NoisePad(714-len(rb))...)
	switch on1 := prev.(type) {
	case *crypt.Layer:
		sess := eng.FindSessionByHeader(on1.ToPriv)
		if sess != nil {
			in := sess.RelayRate *
				lnwire.MilliSatoshi(len(b)) / 2 / 1024 / 1024
			out := sess.RelayRate *
				lnwire.MilliSatoshi(len(rb)) / 2 / 1024 / 1024
			eng.DecSession(sess.ID, in+out, false, "getbalance")
		}
	}
	eng.IterateSessions(func(s *Session) bool {
		if s.ID == on.ID {
			bal = &balance.Layer{
				ID:           on.ID,
				ConfID:       on.ConfID,
				MilliSatoshi: s.Remaining,
			}
			found = true
			return true
		}
		return false
	})
	rb = FormatReply(header,
		Encode(bal), on.Ciphers, on.Nonces)
	rb = append(rb, slice.NoisePad(714-len(rb))...)
	eng.handleMessage(rb, on)
}
