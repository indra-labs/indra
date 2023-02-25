package relay

import (
	"reflect"
	
	"git-indra.lan/indra-labs/indra/pkg/messages/balance"
	"git-indra.lan/indra-labs/indra/pkg/messages/confirm"
	"git-indra.lan/indra-labs/indra/pkg/messages/crypt"
	"git-indra.lan/indra-labs/indra/pkg/messages/delay"
	"git-indra.lan/indra-labs/indra/pkg/messages/exit"
	"git-indra.lan/indra-labs/indra/pkg/messages/forward"
	"git-indra.lan/indra-labs/indra/pkg/messages/getbalance"
	"git-indra.lan/indra-labs/indra/pkg/messages/hiddenservice"
	"git-indra.lan/indra-labs/indra/pkg/messages/response"
	"git-indra.lan/indra-labs/indra/pkg/messages/reverse"
	"git-indra.lan/indra-labs/indra/pkg/messages/session"
	"git-indra.lan/indra-labs/indra/pkg/types"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

func (eng *Engine) handleMessage(b slice.Bytes, prev types.Onion) {
	log.T.F("%v handling received message %v", eng.GetLocalNodeAddress(),
		prev == nil)
	var on1 types.Onion
	var e error
	c := slice.NewCursor()
	if on1, e = Peel(b, c); check(e) {
		return
	}
	switch on := on1.(type) {
	case *balance.Layer:
		if prev == nil {
			log.E.Ln(reflect.TypeOf(on), "requests from outside? absurd!")
			return
		}
		log.T.C(recLog(on, b, eng))
		eng.balance(on, b, c, prev)
	case *confirm.Layer:
		if prev == nil {
			log.E.Ln(reflect.TypeOf(on), "requests from outside? absurd!")
			return
		}
		log.T.C(recLog(on, b, eng))
		eng.confirm(on, b, c, prev)
	case *crypt.Layer:
		log.T.C(recLog(on, b, eng))
		eng.crypt(on, b, c, prev)
	case *delay.Layer:
		if prev == nil {
			log.E.Ln(reflect.TypeOf(on), "requests from outside? absurd!")
			return
		}
		log.T.C(recLog(on, b, eng))
		eng.delay(on, b, c, prev)
	case *exit.Layer:
		if prev == nil {
			log.E.Ln(reflect.TypeOf(on), "requests from outside? absurd!")
			return
		}
		log.T.C(recLog(on, b, eng))
		eng.exit(on, b, c, prev)
	case *forward.Layer:
		if prev == nil {
			log.E.Ln(reflect.TypeOf(on), "requests from outside? absurd!")
			return
		}
		log.T.C(recLog(on, b, eng))
		eng.forward(on, b, c, prev)
	case *getbalance.Layer:
		if prev == nil {
			log.E.Ln(reflect.TypeOf(on), "requests from outside? absurd!")
			return
		}
		log.T.C(recLog(on, b, eng))
		eng.getbalance(on, b, c, prev)
	case *hiddenservice.Layer:
		if prev == nil {
			log.E.Ln(reflect.TypeOf(on), "requests from outside? absurd!")
			return
		}
		log.T.C(recLog(on, b, eng))
		eng.hiddenservice(on, b, c, prev)
	case *response.Layer:
		if prev == nil {
			log.E.Ln(reflect.TypeOf(on), "requests from outside? absurd!")
			return
		}
		log.T.C(recLog(on, b, eng))
		eng.response(on, b, c, prev)
	case *reverse.Layer:
		log.T.C(recLog(on, b, eng))
		eng.reverse(on, b, c, prev)
	case *session.Layer:
		if prev == nil {
			log.E.Ln(reflect.TypeOf(on), "requests from outside? absurd!")
			return
		}
		log.T.C(recLog(on, b, eng))
		eng.session(on, b, c, prev)
	default:
		log.I.S("unrecognised packet", b)
	}
}
