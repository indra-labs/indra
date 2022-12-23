package wire

import (
	"github.com/Indra-Labs/indra"
	"github.com/Indra-Labs/indra/pkg/slice"
	"github.com/Indra-Labs/indra/pkg/types"
	"github.com/Indra-Labs/indra/pkg/wire/cipher"
	"github.com/Indra-Labs/indra/pkg/wire/confirmation"
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
	case cipher.Magic.String():
		var o cipher.Type
		if e = o.Decode(b, c); check(e) {
			return
		}
		on = &o
	case confirmation.Magic.String():
		var o confirmation.Type
		if e = o.Decode(b, c); check(e) {
			return
		}
		on = &o
	case exit.Magic.String():
		var o exit.Type
		if e = o.Decode(b, c); check(e) {
			return
		}
		on = &o
	case forward.Magic.String():
		var o forward.Type
		if e = o.Decode(b, c); check(e) {
			return
		}
		on = &o
	case message.Magic.String():
		var o message.Type
		if e = o.Decode(b, c); check(e) {
			return
		}
		on = &o
	case purchase.Magic.String():
		var o purchase.Type
		if e = o.Decode(b, c); check(e) {
			return
		}
		on = &o
	case reply.Magic.String():
		var o reply.Type
		if e = o.Decode(b, c); check(e) {
			return
		}
		on = &o
	case response.Magic.String():
		var o response.Response
		if e = o.Decode(b, c); check(e) {
			return
		}
		on = o
	case session.Magic.String():
		var o session.Type
		if e = o.Decode(b, c); check(e) {
			return
		}
		on = &o
	case token.Magic.String():
		var o token.Type
		if e = o.Decode(b, c); check(e) {
			return
		}
		on = o
	default:
		return
	}
	return
}
