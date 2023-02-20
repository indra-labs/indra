package relay

//go:generate go run ./pkg/relay/gen/main.go

import (
	"fmt"

	"github.com/davecgh/go-spew/spew"

	"git-indra.lan/indra-labs/indra/pkg/onion/balance"
	"git-indra.lan/indra-labs/indra/pkg/onion/confirm"
	"git-indra.lan/indra-labs/indra/pkg/onion/crypt"
	"git-indra.lan/indra-labs/indra/pkg/onion/delay"
	"git-indra.lan/indra-labs/indra/pkg/onion/dxresponse"
	"git-indra.lan/indra-labs/indra/pkg/onion/exit"
	"git-indra.lan/indra-labs/indra/pkg/onion/forward"
	"git-indra.lan/indra-labs/indra/pkg/onion/getbalance"
	"git-indra.lan/indra-labs/indra/pkg/onion/magicbytes"
	"git-indra.lan/indra-labs/indra/pkg/onion/response"
	"git-indra.lan/indra-labs/indra/pkg/onion/reverse"
	"git-indra.lan/indra-labs/indra/pkg/onion/session"
	"git-indra.lan/indra-labs/indra/pkg/types"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

func Peel(b slice.Bytes, c *slice.Cursor) (on types.Onion, e error) {
	switch b[*c:c.Inc(magicbytes.Len)].String() {
	case balance.MagicString:
		on = &balance.Layer{}
		if e = on.Decode(b, c); check(e) {
			return
		}
	case confirm.MagicString:
		on = &confirm.Layer{}
		if e = on.Decode(b, c); check(e) {
			return
		}
	case crypt.MagicString:
		on = &crypt.Layer{}
		if e = on.Decode(b, c); check(e) {
			return
		}
	case delay.MagicString:
		on = &delay.Layer{}
		if e = on.Decode(b, c); check(e) {
			return
		}
	case dxresponse.MagicString:
		on = &dxresponse.Layer{}
		if e = on.Decode(b, c); check(e) {
			return
		}
	case exit.MagicString:
		on = &exit.Layer{}
		if e = on.Decode(b, c); check(e) {
			return
		}
	case forward.MagicString:
		on = &forward.Layer{}
		if e = on.Decode(b, c); check(e) {
			return
		}
	case getbalance.MagicString:
		on = &getbalance.Layer{}
		if e = on.Decode(b, c); check(e) {
			return
		}
	case response.MagicString:
		on = &response.Layer{}
		if e = on.Decode(b, c); check(e) {
			return
		}
	case reverse.MagicString:
		on = &reverse.Layer{}
		if e = on.Decode(b, c); check(e) {
			return
		}
	case session.MagicString:
		on = &session.Layer{}
		if e = on.Decode(b, c); check(e) {
			return
		}
	default:
		e = fmt.Errorf("message magic not found")
		log.T.C(func() string {
			return fmt.Sprintln(e) + spew.Sdump(b.ToBytes())
		})
		return
	}
	return
}
