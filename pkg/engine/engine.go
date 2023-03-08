package engine

import (
	"reflect"
	
	"github.com/cybriq/qu"
	"go.uber.org/atomic"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/signer"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/util/octet"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

type Engine struct {
	*PendingResponses
	*SessionManager
	*Introductions
	*signer.KeySet
	Load          atomic.Uint32
	TimeoutSignal qu.C
	Pause         qu.C
	ShuttingDown  atomic.Bool
	qu.C
}

type Params struct {
	Tpt             Transport
	IDPrv           *prv.Key
	Node            *Node
	Nodes           []*Node
	NReturnSessions int
}

func NewEngine(p Params) (c *Engine, e error) {
	p.Node.Transport = p.Tpt
	p.Node.IdentityPrv = p.IDPrv
	p.Node.IdentityPub = pub.Derive(p.IDPrv)
	var ks *signer.KeySet
	if _, ks, e = signer.New(); check(e) {
		return
	}
	c = &Engine{
		PendingResponses: &PendingResponses{},
		KeySet:           ks,
		SessionManager:   NewSessionManager(),
		Introductions:    NewIntroductions(),
		TimeoutSignal:    qu.T(),
		Pause:            qu.T(),
		C:                qu.T(),
	}
	c.AddNodes(append([]*Node{p.Node}, p.Nodes...)...)
	// AddIntro a return session for receiving responses, ideally more of these will
	// be generated during operation and rotated out over time.
	for i := 0; i < p.NReturnSessions; i++ {
		c.AddSession(NewSessionData(nonce.NewID(), p.Node, 0, nil, nil, 5))
	}
	return
}

// Start a single thread of the Engine.
func (ng *Engine) Start() {
	log.D.Ln("starting engine")
	for {
		if ng.Handler() {
			break
		}
	}
}

// Cleanup closes and flushes any resources the client opened that require sync
// in order to reopen correctly.
func (ng *Engine) Cleanup() {
	// Do cleanup stuff before shutdown.
}

// Shutdown triggers the shutdown of the client and the Cleanup before
// finishing.
func (ng *Engine) Shutdown() {
	if ng.ShuttingDown.Load() {
		return
	}
	ng.ShuttingDown.Store(true)
	log.T.C(func() string {
		return "shutting down client " + ng.GetLocalNodeAddress().String()
	})
	ng.Cleanup()
	ng.C.Q()
}

func (ng *Engine) HandleMessage(s *octet.Splice, pr Onion) {
	log.T.F("%v handling received message", ng.GetLocalNodeAddress())
	s.SetCursor(0)
	on := Recognise(s)
	log.T.S("bytes", reflect.TypeOf(on), s.GetRange(-1,
		s.GetCursor()).ToBytes(),
		s.GetRange(s.GetCursor(), -1).ToBytes())
	if on != nil {
		if check(on.Decode(s)) {
			return
		}
		if check(on.Handle(s, pr, ng)) {
			log.I.S("unrecognised packet", s.GetRange(-1, -1).ToBytes())
		}
	}
}

func (ng *Engine) Handler() (out bool) {
	log.T.C(func() string {
		return ng.GetLocalNodeAddress().String() +
			" awaiting message"
	})
	var prev Onion
	select {
	case <-ng.C.Wait():
		ng.Shutdown()
		out = true
		break
	case b := <-ng.ReceiveToLocalNode(0):
		s := octet.Load(b, slice.NewCursor())
		ng.HandleMessage(s, prev)
	case p := <-ng.GetLocalNode().PaymentChan.Receive():
		log.D.F("incoming payment for %s: %v", p.ID, p.Amount)
		topUp := false
		ng.IterateSessions(func(s *SessionData) bool {
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
			ng.AddPendingPayment(p)
			log.T.F("awaiting session keys for preimage %x session ID %x",
				p.Preimage, p.ID)
		}
		// For now if we received this we return true. Later this will wait with
		// a timeout on the lnd node returning the success to trigger this.
		p.ConfirmChan <- true
	case <-ng.Pause:
		log.D.Ln("pausing", ng.GetLocalNodeAddress())
		// For testing purposes we need to halt this Handler and discard channel
		// messages.
	out:
		for {
			select {
			case <-ng.GetLocalNode().PaymentChan.Receive():
				log.D.Ln("discarding payments while in pause")
			case <-ng.ReceiveToLocalNode(0):
				log.D.Ln("discarding messages while in pause")
			case <-ng.C.Wait():
				break out
			case <-ng.Pause:
				// This will then resume to the top level select.
				log.D.Ln("unpausing", ng.GetLocalNodeAddress())
				break out
			}
			
		}
	}
	return
}
