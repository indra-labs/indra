package relay

import (
	"git-indra.lan/indra-labs/indra/pkg/messages/intro"
	"git-indra.lan/indra-labs/indra/pkg/types"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

func (eng *Engine) intro(intr *intro.Layer, b slice.Bytes,
	c *slice.Cursor, prev types.Onion) {
	
	if intr.Validate() {
		log.D.F("sending out intro to %s at %s to all known peers",
			intr.Key.ToBase32(), intr.AddrPort.String())
	}
}
