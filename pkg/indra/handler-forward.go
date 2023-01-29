package indra

import (
	"git-indra.lan/indra-labs/lnd/lnd/lnwire"

	"git-indra.lan/indra-labs/indra/pkg/onion/layers/crypt"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/forward"
	"git-indra.lan/indra-labs/indra/pkg/types"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

func (en *Engine) forward(on *forward.Layer, b slice.Bytes,
	c *slice.Cursor, prev types.Onion) {

	// forward the whole buffer received onwards. Usually there will be a
	// crypt.Layer under this which will be unwrapped by the receiver.
	if on.AddrPort.String() == en.Node.AddrPort.String() {
		// it is for us, we want to unwrap the next part.
		en.handleMessage(BudgeUp(b, *c), on)
	} else {
		switch on1 := prev.(type) {
		case *crypt.Layer:
			sess := en.FindSessionByHeader(on1.ToPriv)
			if sess != nil {
				log.D.Ln(on.AddrPort.String(), "forward forward")
				en.DecSession(sess.ID,
					en.RelayRate*lnwire.MilliSatoshi(len(b))/1024/1024)
			}
		}
		// we need to forward this message onion.
		en.Send(on.AddrPort, b)
	}
}
