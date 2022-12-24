package forward

import (
	"net"

	"github.com/Indra-Labs/indra"
	"github.com/Indra-Labs/indra/pkg/slice"
	"github.com/Indra-Labs/indra/pkg/types"
	"github.com/Indra-Labs/indra/pkg/wire/magicbytes"
	log2 "github.com/cybriq/proc/pkg/log"
)

var (
	log                     = log2.GetLogger(indra.PathBase)
	check                   = log.E.Chk
	MagicString             = "fwd"
	Magic                   = slice.Bytes(MagicString)
	MinLen                  = magicbytes.Len + 1 + net.IPv4len
	_           types.Onion = &Type{}
)

// Type forward is just an IP address and a wrapper for another message.
type Type struct {
	net.IP
	types.Onion
}

func (x *Type) Inner() types.Onion   { return x.Onion }
func (x *Type) Insert(o types.Onion) { x.Onion = o }
func (x *Type) Len() int {
	return magicbytes.Len + len(x.IP) + 1 + x.Onion.Len()
}

func (x *Type) Encode(b slice.Bytes, c *slice.Cursor) {
	copy(b[*c:c.Inc(magicbytes.Len)], Magic)
	b[*c] = byte(len(x.IP))
	copy(b[c.Inc(1):c.Inc(len(x.IP))], x.IP)
	x.Onion.Encode(b, c)
}

func (x *Type) Decode(b slice.Bytes, c *slice.Cursor) (e error) {
	if len(b[*c:]) < MinLen {
		return magicbytes.TooShort(len(b), MinLen, string(Magic))
	}
	ipLen := b[*c]
	x.IP = net.IP(b[c.Inc(1):c.Inc(int(ipLen))])
	return
}
