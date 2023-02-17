package reverse

import (
	"net"
	"net/netip"
	
	"git-indra.lan/indra-labs/indra"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/magicbytes"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/types"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

const (
	MagicString = "rv"
	Len         = magicbytes.Len + 1 + net.IPv6len + 2
)

var (
	log               = log2.GetLogger(indra.PathBase)
	check             = log.E.Chk
	Magic             = slice.Bytes(MagicString)
	_     types.Onion = &Layer{}
)

// Layer is reply messages, distinct from forward.Layer messages in that
// the header encryption uses a different secret than the payload. The magic
// bytes signal this to the relay that receives this, which then looks up the
// PayloadHey matching the ToHeaderPub address in the message header. And lastly, each
// step the relay budges up it's message to the front of the packet and puts
// csprng random bytes into the remainder to the same length.
type Layer struct {
	// AddrPort is the address of the next relay in the return leg of a
	// circuit.
	*netip.AddrPort
	types.Onion
}

func (x *Layer) Insert(o types.Onion) { x.Onion = o }
func (x *Layer) Len() int             { return Len + x.Onion.Len() }
func (x *Layer) Encode(b slice.Bytes, c *slice.Cursor) {
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
func (x *Layer) Decode(b slice.Bytes, c *slice.Cursor) (e error) {
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
