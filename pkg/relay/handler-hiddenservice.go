package relay

import (
	"git-indra.lan/indra-labs/indra/pkg/messages/hiddenservice"
	"git-indra.lan/indra-labs/indra/pkg/types"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

func (eng *Engine) hiddenservice(hs *hiddenservice.Layer, b slice.Bytes,
	c *slice.Cursor, prev types.Onion) {
	
	log.D.F("%s adding introduction for key %x", eng.GetLocalNodeAddress(),
		hs.Identity.ToBytes())
	hsk := hs.Identity.ToBytes()
	eng.Introductions.AddIntro(hsk, b[*c:])
	log.I.Ln("stored new introduction, starting broadcast")
	go eng.hiddenserviceBroadcaster(hsk)
}
