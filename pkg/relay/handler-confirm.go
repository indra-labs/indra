package relay

import (
	"git-indra.lan/indra-labs/indra/pkg/onion/confirm"
	"git-indra.lan/indra-labs/indra/pkg/types"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

func (eng *Engine) confirm(on *confirm.Layer,
	b slice.Bytes, c *slice.Cursor, prev types.Onion) {
	
	// When a confirmation arrives check if it is registered for and run the
	// hook that was registered with it.
	eng.PendingResponses.Delete(on.ID, nil)
}
