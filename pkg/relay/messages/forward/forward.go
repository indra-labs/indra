package forward

import (
	"net"
	"net/netip"
	
	"git-indra.lan/indra-labs/indra"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/relay/messages/magicbytes"
	"git-indra.lan/indra-labs/indra/pkg/relay/types"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
	"git-indra.lan/indra-labs/indra/pkg/util/splice"
)

const (
	MagicString = "fw"
	AddrLen     = net.IPv6len + 2
	Len         = magicbytes.Len + 1 + AddrLen
)

var (
	log               = log2.GetLogger(indra.PathBase)
	check             = log.E.Chk
	Magic             = slice.Bytes(MagicString)
	_     types.Onion = &Layer{}
)

// Layer forward is just an IP address and a wrapper for another message.
type Layer struct {
	*netip.AddrPort
	types.Onion
}

// func (x *Layer) String() string {
// 	s, _ := x.AddrPort.MarshalBinary()
// 	return fmt.Sprintf("\n\taddrport: %x %v\n", s, x.AddrPort.String())
// }

func (x *Layer) Insert(o types.Onion) { x.Onion = o }
func (x *Layer) Len() int             { return Len + x.Onion.Len() }
func (x *Layer) Encode(b slice.Bytes, c *slice.Cursor) {
	splice.Splice(b, c).
		Magic(Magic).
		AddrPort(x.AddrPort)
	x.Onion.Encode(b, c)
}
func (x *Layer) Decode(b slice.Bytes, c *slice.Cursor) (e error) {
	if len(b[*c:]) < Len-magicbytes.Len {
		return magicbytes.TooShort(len(b[*c:]), Len-magicbytes.Len, string(Magic))
	}
	splice.Splice(b, c).
		ReadAddrPort(&x.AddrPort)
	return
}
