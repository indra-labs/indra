package intro

import (
	"net"
	"net/netip"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/messages/magicbytes"
	"git-indra.lan/indra-labs/indra/pkg/splice"
	"git-indra.lan/indra-labs/indra/pkg/types"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

const (
	MagicString = "in"
	AddrLen     = net.IPv6len + 3
	Len         = magicbytes.Len + pub.KeyLen + AddrLen
)

var (
	Magic = slice.Bytes(MagicString)
)

type Layer struct {
	*pub.Key
	*netip.AddrPort
}

func (im *Layer) Insert(o types.Onion) {}
func (im *Layer) Len() int             { return Len }

func (im *Layer) Encode(b slice.Bytes, c *slice.Cursor) {
	splice.Splice(b, c).Magic(Magic).Pubkey(im.Key).AddrPort(im.AddrPort)
	return
}

func (im *Layer) Decode(b slice.Bytes, c *slice.Cursor) (e error) {
	splice.Splice(b, c).ReadPubkey(&im.Key).ReadAddrPort(&im.AddrPort)
	return
}
