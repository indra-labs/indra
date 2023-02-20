package relay

import (
	"git-indra.lan/indra-labs/lnd/lnd/lnwire"

	"git-indra.lan/indra-labs/indra/pkg/onion/crypt"
	"git-indra.lan/indra-labs/indra/pkg/onion/forward"
	"git-indra.lan/indra-labs/indra/pkg/types"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

func (eng *Engine) forward(on *forward.Layer, b slice.Bytes,
	c *slice.Cursor, prev types.Onion) {

	// forward the whole buffer received onwards. Usually there will be a
	// crypt.Layer under this which will be unwrapped by the receiver.
	if on.AddrPort.String() == eng.GetLocalNodeAddress().String() {
		// it is for us, we want to unwrap the next part.
		eng.handleMessage(BudgeUp(b, *c), on)
	} else {
		switch on1 := prev.(type) {
		case *crypt.Layer:
			sess := eng.FindSessionByHeader(on1.ToPriv)
			if sess != nil {
				eng.DecSession(sess.ID,
					eng.GetLocalNodeRelayRate()*lnwire.MilliSatoshi(len(b))/1024/1024,
					false, "forward")
			}
		}
		// we need to forward this message onion.
		eng.Send(on.AddrPort, BudgeUp(b, *c))
	}
}
