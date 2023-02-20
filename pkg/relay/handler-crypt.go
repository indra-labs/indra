package relay

import (
	"git-indra.lan/indra-labs/indra/pkg/onion/crypt"
	"git-indra.lan/indra-labs/indra/pkg/onion/magicbytes"
	"git-indra.lan/indra-labs/indra/pkg/onion/session"
	"git-indra.lan/indra-labs/indra/pkg/types"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

func (eng *Engine) crypt(on *crypt.Layer, b slice.Bytes,
	c *slice.Cursor, prev types.Onion) {

	// this is probably an encrypted crypt for us.
	hdr, _, _, identity := eng.FindCloaked(on.Cloak)
	if hdr == nil {
		log.T.Ln("no matching key found from cloaked key")
		return
	}
	on.ToPriv = hdr
	on.Decrypt(hdr, b, c)
	if identity {
		if string(b[*c:][:magicbytes.Len]) != session.MagicString {
			log.T.Ln("dropping message due to identity key with" +
				" no following session")
			return
		}
		eng.handleMessage(BudgeUp(b, *c), on)
		return
	}
	eng.handleMessage(BudgeUp(b, *c), on)
}
