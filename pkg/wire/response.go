package wire

import (
	"github.com/Indra-Labs/indra/pkg/slice"
	"github.com/Indra-Labs/indra/pkg/types"
)

// Response messages are what are carried back via Return messages from an Exit.
type Response slice.Bytes

var (
	ResponseMagic             = slice.Bytes("res")
	_             types.Onion = Response{}
)

func (x Response) Inner() types.Onion   { return nil }
func (x Response) Insert(_ types.Onion) {}
func (x Response) Len() int             { return MagicLen + len(x) + 4 }

func (x Response) Encode(o slice.Bytes, c *slice.Cursor) {
	copy(o[*c:c.Inc(MagicLen)], ResponseMagic)
	bytesLen := slice.NewUint32()
	slice.EncodeUint32(bytesLen, len(x))
	copy(o[*c:c.Inc(slice.Uint32Len)], bytesLen)
	copy(o[*c:c.Inc(len(x))], x)
}

func (x Response) Decode(b slice.Bytes, c *slice.Cursor) (in interface{},
	e error) {

	magic := ResponseMagic
	if !CheckMagic(b, magic) {
		return ReturnError(ErrWrongMagic, x, b, magic)
	}

	return
}
