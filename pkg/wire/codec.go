package wire

import (
	"github.com/Indra-Labs/indra"
	"github.com/Indra-Labs/indra/pkg/slice"
	"github.com/Indra-Labs/indra/pkg/types"
	"github.com/Indra-Labs/indra/pkg/wire/cipher"
	"github.com/Indra-Labs/indra/pkg/wire/confirmation"
	"github.com/Indra-Labs/indra/pkg/wire/delay"
	"github.com/Indra-Labs/indra/pkg/wire/exit"
	"github.com/Indra-Labs/indra/pkg/wire/forward"
	"github.com/Indra-Labs/indra/pkg/wire/magicbytes"
	"github.com/Indra-Labs/indra/pkg/wire/message"
	"github.com/Indra-Labs/indra/pkg/wire/purchase"
	"github.com/Indra-Labs/indra/pkg/wire/reply"
	"github.com/Indra-Labs/indra/pkg/wire/response"
	"github.com/Indra-Labs/indra/pkg/wire/session"
	"github.com/Indra-Labs/indra/pkg/wire/token"
	log2 "github.com/cybriq/proc/pkg/log"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

func EncodeOnion(on types.Onion) (b slice.Bytes) {
	b = make(slice.Bytes, on.Len())
	var sc slice.Cursor
	c := &sc
	on.Encode(b, c)
	return
}

func PeelOnion(b slice.Bytes, c *slice.Cursor) (on types.Onion, e error) {
	switch b[*c:c.Inc(magicbytes.Len)].String() {
	case cipher.MagicString:
		var o cipher.OnionSkin
		if e = o.Decode(b, c); check(e) {
			return
		}
		on = &o
	case confirmation.MagicString:
		var o confirmation.OnionSkin
		if e = o.Decode(b, c); check(e) {
			return
		}
		on = &o
	case delay.MagicString:
		var o delay.OnionSkin
		if e = o.Decode(b, c); check(e) {
			return
		}
		on = &o
	case exit.MagicString:
		var o exit.OnionSkin
		if e = o.Decode(b, c); check(e) {
			return
		}
		on = &o
	case forward.MagicString:
		var o forward.OnionSkin
		if e = o.Decode(b, c); check(e) {
			return
		}
		on = &o
	case message.MagicString:
		var o message.OnionSkin
		if e = o.Decode(b, c); check(e) {
			return
		}
		on = &o
	case purchase.MagicString:
		var o purchase.OnionSkin
		if e = o.Decode(b, c); check(e) {
			return
		}
		on = &o
	case reply.MagicString:
		var o reply.OnionSkin
		if e = o.Decode(b, c); check(e) {
			return
		}
		on = &o
	case response.MagicString:
		var o response.OnionSkin
		if e = o.Decode(b, c); check(e) {
			return
		}
		on = o
	case session.MagicString:
		var o session.OnionSkin
		if e = o.Decode(b, c); check(e) {
			return
		}
		on = &o
	case token.MagicString:
		var o token.OnionSkin
		if e = o.Decode(b, c); check(e) {
			return
		}
		on = o
	default:
		return
	}
	return
}
