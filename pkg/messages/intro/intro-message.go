package intro

import (
	"net"
	"net/netip"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/messages/magicbytes"
	"git-indra.lan/indra-labs/indra/pkg/splice"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

const (
	MagicString = "in"
	AddrLen     = net.IPv6len + 2
	Len         = magicbytes.Len + pub.KeyLen + AddrLen
)

var (
	Magic = slice.Bytes(MagicString)
)

type IntroductionMessage struct {
	*pub.Key
	*netip.AddrPort
}

func Encode(im *IntroductionMessage) (o slice.Bytes) {
	b, c := make(slice.Bytes, Len), slice.NewCursor()
	splice.Splice(b, c).Magic(Magic).Pubkey(im.Key).AddrPort(im.AddrPort)
	return
}

func Decode(b slice.Bytes) (o *IntroductionMessage) {
	o = &IntroductionMessage{}
	c := slice.NewCursor()
	c.Inc(magicbytes.Len)
	splice.Splice(b, c).ReadPubkey(&o.Key).ReadAddrPort(&o.AddrPort)
	return
}
