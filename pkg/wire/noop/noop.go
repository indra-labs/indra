package noop

import (
	"github.com/indra-labs/indra/pkg/types"
	"github.com/indra-labs/indra/pkg/util/slice"
)

type OnionSkin struct {
}

func (x *OnionSkin) Inner() types.Onion   { return nil }
func (x *OnionSkin) Insert(o types.Onion) {}
func (x *OnionSkin) Len() int             { return 0 }
func (x *OnionSkin) Encode(b slice.Bytes,
	c *slice.Cursor) {
}
func (x *OnionSkin) Decode(b slice.Bytes,
	c *slice.Cursor) (e error) {
	return
}
