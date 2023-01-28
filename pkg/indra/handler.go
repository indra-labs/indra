package indra

import (
	"fmt"
	"reflect"

	"github.com/davecgh/go-spew/spew"

	"github.com/indra-labs/indra/pkg/onion"
	"github.com/indra-labs/indra/pkg/onion/layers/balance"
	"github.com/indra-labs/indra/pkg/onion/layers/confirm"
	"github.com/indra-labs/indra/pkg/onion/layers/crypt"
	"github.com/indra-labs/indra/pkg/onion/layers/delay"
	"github.com/indra-labs/indra/pkg/onion/layers/exit"
	"github.com/indra-labs/indra/pkg/onion/layers/forward"
	"github.com/indra-labs/indra/pkg/onion/layers/getbalance"
	"github.com/indra-labs/indra/pkg/onion/layers/response"
	"github.com/indra-labs/indra/pkg/onion/layers/reverse"
	"github.com/indra-labs/indra/pkg/onion/layers/session"
	"github.com/indra-labs/indra/pkg/traffic"
	"github.com/indra-labs/indra/pkg/types"
	"github.com/indra-labs/indra/pkg/util/slice"
)

func (cl *Engine) handler() (out bool) {
	log.T.C(func() string {
		return cl.AddrPort.String() +
			" awaiting message"
	})
	var prev types.Onion
	select {
	case <-cl.C.Wait():
		cl.Cleanup()
		out = true
		break
	case b := <-cl.Node.Receive():
		cl.handleMessage(b, prev)
	case p := <-cl.PaymentChan:
		log.T.S("incoming payment", cl.AddrPort.String(), p)
		topUp := false
		cl.IterateSessions(func(s *traffic.Session) bool {
			if s.Preimage == p.Preimage {
				s.IncSats(p.Amount)
				topUp = true
				log.T.F("topping up %x with %d mSat",
					s.ID, p.Amount)
				return true
			}
			return false
		})
		if !topUp {
			cl.AddPendingPayment(p)
			log.T.F("awaiting session keys for preimage %x",
				p.Preimage)
		}
	}
	return
}

func (cl *Engine) handleMessage(b slice.Bytes, prev types.Onion) {
	// process received message
	var on types.Onion
	var e error
	c := slice.NewCursor()
	if on, e = onion.Peel(b, c); check(e) {
		return
	}
	switch on := on.(type) {
	case *balance.Layer:
		log.T.C(recLog(on, b, cl))
		cl.balance(on, b, c, prev)
	case *confirm.Layer:
		log.T.C(recLog(on, b, cl))
		cl.confirm(on, b, c, prev)
	case *crypt.Layer:
		log.T.C(recLog(on, b, cl))
		cl.crypt(on, b, c, prev)
	case *delay.Layer:
		log.T.C(recLog(on, b, cl))
		cl.delay(on, b, c, prev)
	case *exit.Layer:
		log.T.C(recLog(on, b, cl))
		cl.exit(on, b, c, prev)
	case *forward.Layer:
		log.T.C(recLog(on, b, cl))
		cl.forward(on, b, c, prev)
	case *getbalance.Layer:
		log.T.C(recLog(on, b, cl))
		cl.getBalance(on, b, c, prev)
	case *reverse.Layer:
		log.T.C(recLog(on, b, cl))
		cl.reverse(on, b, c, prev)
	case *response.Layer:
		log.T.C(recLog(on, b, cl))
		cl.response(on, b, c, prev)
	case *session.Layer:
		log.T.C(recLog(on, b, cl))
		cl.session(on, b, c, prev)
	default:
		log.I.S("unrecognised packet", b)
	}
}

// utility functions

func recLog(on types.Onion, b slice.Bytes, cl *Engine) func() string {
	return func() string {
		return cl.AddrPort.String() +
			" received " +
			fmt.Sprint(reflect.TypeOf(on)) + "\n" +
			spew.Sdump(b.ToBytes())
	}
}
