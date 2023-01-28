package indra

import (
	"github.com/indra-labs/indra/pkg/onion/layers/confirm"
	"github.com/indra-labs/indra/pkg/types"
	"github.com/indra-labs/indra/pkg/util/slice"
)

func (en *Engine) confirm(on *confirm.Layer,
	b slice.Bytes, c *slice.Cursor, prev types.Onion) {

	// When a confirm arrives check if it is registered for and run
	// the hook that was registered with it.
	log.T.S(en.Confirms)
	en.Confirms.Confirm(on.ID)
}
