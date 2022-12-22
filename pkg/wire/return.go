package wire

import (
	"net"

	"github.com/Indra-Labs/indra/pkg/slice"
	"github.com/Indra-Labs/indra/pkg/types"
)

// Return messages are distinct from Forward messages in that the header
// encryption uses a different secret than the payload. The magic bytes signal
// this to the relay that receives this, which then looks up the Return key
// matching the To address in the message header.
type Return struct {
	// IP is the address of the next relay in the return leg of a circuit.
	net.IP
	types.Onion
}

var (
	ReturnMagic             = slice.Bytes("rtn")
	_           types.Onion = &Return{}
)

func (x *Return) Inner() types.Onion   { return x.Onion }
func (x *Return) Insert(o types.Onion) { x.Onion = o }
func (x *Return) Len() int {
	return MagicLen + len(x.IP) + 1 + x.Onion.Len()
}

func (x *Return) Encode(o slice.Bytes, c *slice.Cursor) {
	copy(o[*c:c.Inc(MagicLen)], ReturnMagic)
	o[*c] = byte(len(x.IP))
	copy(o[c.Inc(1):c.Inc(len(x.IP))], x.IP)
	x.Onion.Encode(o, c)
}

func (x *Return) Decode(b slice.Bytes, c *slice.Cursor) (in interface{},
	e error) {

	magic := ReturnMagic
	if !CheckMagic(b, magic) {
		return ReturnError(ErrWrongMagic, x, b, magic)
	}

	return
}
