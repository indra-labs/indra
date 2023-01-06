package reverse

import (
	"net"
	"net/netip"

	"github.com/indra-labs/indra"
	log2 "github.com/indra-labs/indra/pkg/log"
	"github.com/indra-labs/indra/pkg/slice"
	"github.com/indra-labs/indra/pkg/types"
	"github.com/indra-labs/indra/pkg/wire/magicbytes"
)

const (
	MagicString = "rv"
	Len         = magicbytes.Len + 1 + net.IPv6len + 2
)

var (
	log               = log2.GetLogger(indra.PathBase)
	check             = log.E.Chk
	Magic             = slice.Bytes(MagicString)
	_     types.Onion = &OnionSkin{}
)

// OnionSkin is reply messages, distinct from forward.OnionSkin messages in that
// the header encryption uses a different secret than the payload. The magic
// bytes signal this to the relay that receives this, which then looks up the
// PayloadHey matching the To address in the message header. And lastly, each
// step the relay budges up it's message to the front of the packet and puts
// csprng random bytes into the remainder to the same length.
type OnionSkin struct {
	// AddrPort is the address of the next relay in the return leg of a
	// circuit.
	*netip.AddrPort
	types.Onion
}

func (x *OnionSkin) Inner() types.Onion   { return x.Onion }
func (x *OnionSkin) Insert(o types.Onion) { x.Onion = o }
func (x *OnionSkin) Len() int             { return Len + x.Onion.Len() }

func (x *OnionSkin) Encode(b slice.Bytes, c *slice.Cursor) {
	copy(b[*c:c.Inc(magicbytes.Len)], Magic)
	var ap []byte
	var e error
	if ap, e = x.AddrPort.MarshalBinary(); check(e) {
		return
	}
	b[*c] = byte(len(ap))
	copy(b[c.Inc(1):c.Inc(Len-magicbytes.Len-1)], ap)
	x.Onion.Encode(b, c)
}

func (x *OnionSkin) Decode(b slice.Bytes, c *slice.Cursor) (e error) {
	if len(b[*c:]) < Len-magicbytes.Len {
		return magicbytes.TooShort(len(b[*c:]), Len-magicbytes.Len, string(Magic))
	}
	apLen := b[*c]
	apBytes := b[c.Inc(1):c.Inc(Len-magicbytes.Len-1)]
	x.AddrPort = &netip.AddrPort{}
	if e = x.AddrPort.UnmarshalBinary(apBytes[:apLen]); check(e) {
		return
	}
	return
}
