package relay

import (
	"fmt"
	"reflect"
	
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
		eng.Shutdown()
		out = true
		break
	case b := <-eng.ReceiveToLocalNode(0):
		eng.handleMessage(b, prev)
	case p := <-eng.GetLocalNode().PaymentChan.Receive():
		log.D.F("incoming payment for %s: %v", p.ID, p.Amount)
		topUp := false
		eng.IterateSessions(func(s *Session) bool {
			if s.Preimage == p.Preimage {
				s.IncSats(p.Amount, false, "top-up")
				topUp = true
				log.T.F("topping up %x with %v",
					s.ID, p.Amount)
				return true
			}
			return false
		})
		if !topUp {
			eng.AddPendingPayment(p)
			log.T.F("awaiting session keys for preimage %x session ID %x",
				p.Preimage, p.ID)
		}
		// For now if we received this we return true.
		// Later this will wait with a timeout on the lnd node returning the
		// success to trigger this.
		p.ConfirmChan <- true
	case <-eng.Pause:
		log.D.Ln("pausing", eng.GetLocalNodeAddress())
		// For testing purposes we need to halt this handler and discard
		// channel messages.
	out:
		for {
			select {
			case <-eng.GetLocalNode().PaymentChan.Receive():
				log.D.Ln("discarding payments while in pause")
			case <-eng.ReceiveToLocalNode(0):
				log.D.Ln("discarding messages while in pause")
			case <-eng.C.Wait():
				break out
			case <-eng.Pause:
				// This will then resume to the top level select.
				log.D.Ln("unpausing", eng.GetLocalNodeAddress())
				break out
			}
			
		}
	}
	return
}

// utility functions

func recLog(on types.Onion, b slice.Bytes, cl *Engine) func() string {
	return func() string {
		return cl.GetLocalNodeAddress().String() +
			" received " +
			fmt.Sprint(reflect.TypeOf(on)) + "\n" +
			""
		// spew.Sdump(b.ToBytes())
	}
}
