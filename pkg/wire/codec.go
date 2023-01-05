package wire

import (
	"fmt"

	"github.com/Indra-Labs/indra"
	log2 "github.com/Indra-Labs/indra/pkg/log"
	"github.com/Indra-Labs/indra/pkg/slice"
	"github.com/Indra-Labs/indra/pkg/types"
	"github.com/Indra-Labs/indra/pkg/wire/cipher"
	"github.com/Indra-Labs/indra/pkg/wire/confirm"
	"github.com/Indra-Labs/indra/pkg/wire/delay"
	"github.com/Indra-Labs/indra/pkg/wire/exit"
	"github.com/Indra-Labs/indra/pkg/wire/forward"
	"github.com/Indra-Labs/indra/pkg/wire/layer"
	"github.com/Indra-Labs/indra/pkg/wire/magicbytes"
	"github.com/Indra-Labs/indra/pkg/wire/purchase"
	"github.com/Indra-Labs/indra/pkg/wire/response"
	"github.com/Indra-Labs/indra/pkg/wire/reverse"
	"github.com/Indra-Labs/indra/pkg/wire/session"
	"github.com/Indra-Labs/indra/pkg/wire/token"
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
		o := &cipher.OnionSkin{}
		if e = o.Decode(b, c); check(e) {
			return
		}
		on = o
	case confirm.MagicString:
		o := &confirm.OnionSkin{}
		if e = o.Decode(b, c); check(e) {
			return
		}
		on = o
	case delay.MagicString:
		o := &delay.OnionSkin{}
		if e = o.Decode(b, c); check(e) {
			return
		}
		on = o
	case exit.MagicString:
		o := &exit.OnionSkin{}
		if e = o.Decode(b, c); check(e) {
			return
		}
		on = o
	case forward.MagicString:
		o := &forward.OnionSkin{}
		if e = o.Decode(b, c); check(e) {
			return
		}
		on = o
	case layer.MagicString:
		var o layer.OnionSkin
		if e = o.Decode(b, c); check(e) {
			return
		}
		on = &o
	case purchase.MagicString:
		o := &purchase.OnionSkin{}
		if e = o.Decode(b, c); check(e) {
			return
		}
		on = o
	case reverse.MagicString:
		o := &reverse.OnionSkin{}
		if e = o.Decode(b, c); check(e) {
			return
		}
		on = o
	case response.MagicString:
		o := response.New()
		if e = o.Decode(b, c); check(e) {
			return
		}
		on = o
	case session.MagicString:
		o := &session.OnionSkin{}
		if e = o.Decode(b, c); check(e) {
			return
		}
		on = o
	case token.MagicString:
		o := token.NewOnionSkin()
		if e = o.Decode(b, c); check(e) {
			return
		}
		on = o
	default:
		e = fmt.Errorf("message magic not found")
		check(e)
		log.I.S(b.ToBytes())
		return
	}
	return
}
