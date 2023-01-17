package noop

import (
	"github.com/indra-labs/indra/pkg/types"
	"github.com/indra-labs/indra/pkg/util/slice"
)

type Layer struct {
}

func (x *Layer) Inner() types.Onion   { return nil }
func (x *Layer) Insert(o types.Onion) {}
func (x *Layer) Len() int             { return 0 }
func (x *Layer) Encode(b slice.Bytes,
	c *slice.Cursor) {
}
func (x *Layer) Decode(b slice.Bytes,
	c *slice.Cursor) (e error) {
	return
}
