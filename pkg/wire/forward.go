package wire

import (
	"net"

	"github.com/Indra-Labs/indra/pkg/slice"
	"github.com/Indra-Labs/indra/pkg/types"
)

// Forward is just an IP address and a wrapper for another message.
type Forward struct {
	net.IP
	types.Onion
}

var (
	ForwardMagic             = slice.Bytes("fwd")
	_            types.Onion = &Forward{}
)

func (x *Forward) Inner() types.Onion   { return x.Onion }
func (x *Forward) Insert(o types.Onion) { x.Onion = o }
func (x *Forward) Len() int {
	return MagicLen + len(x.IP) + 1 + x.Onion.Len()
}

func (x *Forward) Encode(o slice.Bytes, c *slice.Cursor) {
	copy(o[*c:c.Inc(MagicLen)], ForwardMagic)
	o[*c] = byte(len(x.IP))
	copy(o[c.Inc(1):c.Inc(len(x.IP))], x.IP)
	x.Onion.Encode(o, c)
}

func (x *Forward) Decode(b slice.Bytes, c *slice.Cursor) (in interface{},
	e error) {

	magic := ForwardMagic
	if !CheckMagic(b, magic) {
		return ReturnError(ErrWrongMagic, x, b, magic)
	}

	return
}
