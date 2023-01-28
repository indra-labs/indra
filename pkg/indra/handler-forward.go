package indra

import (
	"github.com/indra-labs/lnd/lnd/lnwire"

	"github.com/indra-labs/indra/pkg/onion/layers/crypt"
	"github.com/indra-labs/indra/pkg/onion/layers/forward"
	"github.com/indra-labs/indra/pkg/types"
	"github.com/indra-labs/indra/pkg/util/slice"
)

func (cl *Engine) forward(on *forward.Layer, b slice.Bytes,
	c *slice.Cursor, prev types.Onion) {

	// forward the whole buffer received onwards. Usually there will be a
	// crypt.Layer under this which will be unwrapped by the receiver.
	if on.AddrPort.String() == cl.Node.AddrPort.String() {
		// it is for us, we want to unwrap the next part.
		cl.handleMessage(BudgeUp(b, *c), on)
	} else {
		switch on1 := prev.(type) {
		case *crypt.Layer:
			sess := cl.FindSessionByHeader(on1.ToPriv)
			if sess != nil {
				log.D.Ln(on.AddrPort.String(), "forward forward")
				cl.DecSession(sess.ID,
					cl.RelayRate*lnwire.MilliSatoshi(len(b))/1024/1024)
			}
		}
		// we need to forward this message onion.
		cl.Send(on.AddrPort, b)
	}
}
