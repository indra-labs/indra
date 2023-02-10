package relay

import (
	"fmt"
	"reflect"

	"github.com/davecgh/go-spew/spew"

	"git-indra.lan/indra-labs/indra/pkg/onion"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/balance"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/confirm"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/crypt"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/delay"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/exit"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/forward"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/getbalance"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/response"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/reverse"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/session"
	"git-indra.lan/indra-labs/indra/pkg/traffic"
	"git-indra.lan/indra-labs/indra/pkg/types"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

func (eng *Engine) handler() (out bool) {
	log.T.C(func() string {
		return eng.GetLocalNodeAddress().String() +
			" awaiting message"
	})
	var prev types.Onion
	select {
	case <-eng.C.Wait():
		eng.Cleanup()
		out = true
		break
	case b := <-eng.ReceiveToLocalNode(0):
		eng.handleMessage(b, prev)
	case p := <-eng.PaymentChan:
		log.D.F("incoming payment for %x: %v", p.ID, p.Amount)
		topUp := false
		eng.IterateSessions(func(s *traffic.Session) bool {
			log.D.F("session preimage %x payment preimage %x", s.Preimage,
				p.Preimage)
			if s.Preimage == p.Preimage {
				s.IncSats(p.Amount, false, "top-up")
				topUp = true
				log.T.F("topping up %x with %d mSat",
					s.ID, p.Amount)
				return true
			}
			return false
		})
		if !topUp {
			eng.AddPendingPayment(p)
			log.T.F("awaiting session keys for preimage %x",
				p.Preimage)
		}
	}
	return
}

func (eng *Engine) handleMessage(b slice.Bytes, prev types.Onion) {
	log.T.Ln("process received message")
	var on types.Onion
	var e error
	c := slice.NewCursor()
	if on, e = onion.Peel(b, c); check(e) {
		return
	}
	switch on := on.(type) {
	case *balance.Layer:
		log.T.C(recLog(on, b, eng))
		eng.balance(on, b, c, prev)
	case *confirm.Layer:
		log.T.C(recLog(on, b, eng))
		eng.confirm(on, b, c, prev)
	case *crypt.Layer:
		log.T.C(recLog(on, b, eng))
		eng.crypt(on, b, c, prev)
	case *delay.Layer:
		log.T.C(recLog(on, b, eng))
		eng.delay(on, b, c, prev)
	case *exit.Layer:
		log.T.C(recLog(on, b, eng))
		eng.exit(on, b, c, prev)
	case *forward.Layer:
		log.T.C(recLog(on, b, eng))
		eng.forward(on, b, c, prev)
	case *getbalance.Layer:
		log.T.C(recLog(on, b, eng))
		eng.getBalance(on, b, c, prev)
	case *reverse.Layer:
		log.T.C(recLog(on, b, eng))
		eng.reverse(on, b, c, prev)
	case *response.Layer:
		log.T.C(recLog(on, b, eng))
		eng.response(on, b, c, prev)
	case *session.Layer:
		log.T.C(recLog(on, b, eng))
		eng.session(on, b, c, prev)
	default:
		log.I.S("unrecognised packet", b)
	}
}

// utility functions

func recLog(on types.Onion, b slice.Bytes, cl *Engine) func() string {
	return func() string {
		return cl.GetLocalNodeAddress().String() +
			" received " +
			fmt.Sprint(reflect.TypeOf(on)) + "\n" +
			spew.Sdump(b.ToBytes())
	}
}
