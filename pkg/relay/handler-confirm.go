package relay

import (
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/confirm"
	"git-indra.lan/indra-labs/indra/pkg/types"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

func (en *Engine) confirm(on *confirm.Layer,
	b slice.Bytes, c *slice.Cursor, prev types.Onion) {

	log.T.F("processing confirmation %x", on.ID)
	// When a confirmation arrives check if it is registered for and run the
	// hook that was registered with it.
	en.Pending.Delete(on.ID, nil)
}
