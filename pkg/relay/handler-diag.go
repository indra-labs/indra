package relay

import (
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/onion"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/crypt"
	diag "git-indra.lan/indra-labs/indra/pkg/onion/layers/diagnostic"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/dxresponse"
	"git-indra.lan/indra-labs/indra/pkg/types"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

func (eng *Engine) diag(dx *diag.Layer, b slice.Bytes,
	c *slice.Cursor, prev types.Onion) {
	
	// A diagnostic message requires a response containing the session ID it
	// is received on and the relay load level.
	var sessID nonce.ID
	switch on := prev.(type) {
	case *crypt.Layer:
		sess := eng.FindSessionByHeader(on.ToPriv)
		if sess == nil {
			return
		}
		sessID = sess.ID
	}
	eng.Lock()
	res := onion.Encode(&dxresponse.Layer{
		ID:   sessID,
		Load: eng.Load,
	})
	eng.Unlock()
	rb := FormatReply(b[*c:c.Inc(crypt.ReverseHeaderLen)],
		res, dx.Ciphers, dx.Nonces)
	// Unlike the exit message, this response is not the only message that has
	// to be forwarded. This reply header needs to be processed by forwarding to
	// the first hop specified in the header, and the rest of the message also,
	// as it has a forward message at the front of it.
	eng.handleMessage(rb, dx)
	eng.handleMessage(BudgeUp(b, *c), dx)
}
