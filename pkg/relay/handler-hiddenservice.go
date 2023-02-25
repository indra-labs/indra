package relay

import (
	"git-indra.lan/indra-labs/indra/pkg/onion/hiddenservice"
	"git-indra.lan/indra-labs/indra/pkg/types"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

func (eng *Engine) hiddenservice(hs *hiddenservice.Layer, b slice.Bytes,
	c *slice.Cursor, prev types.Onion) {
	
	log.D.F("%s adding introduction for key %x", eng.GetLocalNodeAddress(),
		hs.Identity.ToBytes())
	eng.Introductions.Add(hs.Identity.ToBytes(), b[*c:])
	log.I.Ln("stored new introduction")
}
