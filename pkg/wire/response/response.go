package response

import (
	"github.com/Indra-Labs/indra"
	"github.com/Indra-Labs/indra/pkg/slice"
	"github.com/Indra-Labs/indra/pkg/types"
	"github.com/Indra-Labs/indra/pkg/wire/magicbytes"
	log2 "github.com/cybriq/proc/pkg/log"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

// Response messages are what are carried back via Return messages from an Exit.
type Response slice.Bytes

var (
	Magic              = slice.Bytes("res")
	MinLen             = magicbytes.Len + slice.Uint32Len
	_      types.Onion = Response{}
)

func (x Response) Inner() types.Onion   { return nil }
func (x Response) Insert(_ types.Onion) {}
func (x Response) Len() int             { return MinLen + len(x) }

func (x Response) Encode(o slice.Bytes, c *slice.Cursor) {
	copy(o[*c:c.Inc(magicbytes.Len)], Magic)
	bytesLen := slice.NewUint32()
	slice.EncodeUint32(bytesLen, len(x))
	copy(o[*c:c.Inc(slice.Uint32Len)], bytesLen)
	copy(o[*c:c.Inc(len(x))], x)
}

func (x Response) Decode(b slice.Bytes) (e error) {

	if !magicbytes.CheckMagic(b, Magic) {
		return magicbytes.WrongMagic(x, b, Magic)
	}
	if len(b) < MinLen {
		return magicbytes.TooShort(len(b), MinLen, string(Magic))
	}
	sc := slice.Cursor(0)
	c := &sc
	_ = c

	return
}
