package indra

import (
	"fmt"

	"git-indra.lan/indra-labs/lnd/lnd/lnwire"

	"git-indra.lan/indra-labs/indra/pkg/onion"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/balance"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/crypt"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/getbalance"
	"git-indra.lan/indra-labs/indra/pkg/traffic"
	"git-indra.lan/indra-labs/indra/pkg/types"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

func (en *Engine) getBalance(on *getbalance.Layer,
	b slice.Bytes, c *slice.Cursor, prev types.Onion) {

	log.T.S(on)
	var found bool
	var bal *balance.Layer
	en.IterateSessions(func(s *traffic.Session) bool {
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
		fmt.Println("session not found")
		return
	}
	rb := FormatReply(b[*c:c.Inc(crypt.ReverseHeaderLen)],
		onion.Encode(bal), on.Ciphers, on.Nonces)
	rb = append(rb, slice.NoisePad(714-len(rb))...)
	switch on1 := prev.(type) {
	case *crypt.Layer:
		sess := en.FindSessionByHeader(on1.ToPriv)
		if sess != nil {
			log.D.Ln("getbalance reply")
			in := sess.RelayRate *
				lnwire.MilliSatoshi(len(b)) / 2 / 1024 / 1024
			out := sess.RelayRate *
				lnwire.MilliSatoshi(len(rb)) / 2 / 1024 / 1024
			en.DecSession(sess.ID, in+out)
		}
	}
	en.handleMessage(rb, on)
}
