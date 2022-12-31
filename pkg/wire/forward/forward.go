package forward

import (
	"fmt"
	"net"
	"net/netip"

	"github.com/Indra-Labs/indra"
	"github.com/Indra-Labs/indra/pkg/slice"
	"github.com/Indra-Labs/indra/pkg/types"
	"github.com/Indra-Labs/indra/pkg/wire/magicbytes"
	log2 "github.com/cybriq/proc/pkg/log"
)

const (
	MagicString = "fw"
	Len         = magicbytes.Len + 1 + net.IPv6len + 2
)

var (
	log               = log2.GetLogger(indra.PathBase)
	check             = log.E.Chk
	Magic             = slice.Bytes(MagicString)
	_     types.Onion = &OnionSkin{}
)

// OnionSkin forward is just an IP address and a wrapper for another message.
type OnionSkin struct {
	*netip.AddrPort
	types.Onion
}

func (x *OnionSkin) String() string {
	s, _ := x.AddrPort.MarshalBinary()
	return fmt.Sprintf("\n\taddrport: %x %v\n", s, x.AddrPort.String())
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
