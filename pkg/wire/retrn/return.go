package retrn

import (
	"net"

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

// Type is return messages, distinct from Forward messages in that the header
// encryption uses a different secret than the payload. The magic bytes signal
// this to the relay that receives this, which then looks up the PayloadHey
// matching the To address in the message header.
type Type struct {
	// IP is the address of the next relay in the return leg of a circuit.
	net.IP
	types.Onion
}

var (
	Magic              = slice.Bytes("rtn")
	MinLen             = magicbytes.Len + 1 + net.IPv4len
	_      types.Onion = &Type{}
)

func (x *Type) Inner() types.Onion   { return x.Onion }
func (x *Type) Insert(o types.Onion) { x.Onion = o }
func (x *Type) Len() int {
	return magicbytes.Len + len(x.IP) + 1 + x.Onion.Len()
}

func (x *Type) Encode(o slice.Bytes, c *slice.Cursor) {
	copy(o[*c:c.Inc(magicbytes.Len)], Magic)
	o[*c] = byte(len(x.IP))
	copy(o[c.Inc(1):c.Inc(len(x.IP))], x.IP)
	x.Onion.Encode(o, c)
}

func (x *Type) Decode(b slice.Bytes, c *slice.Cursor) (e error) {

	if !magicbytes.CheckMagic(b, Magic) {
		return magicbytes.WrongMagic(x, b, Magic)
	}
	if len(b) < MinLen {
		return magicbytes.TooShort(len(b), MinLen, string(Magic))
	}

	return
}
