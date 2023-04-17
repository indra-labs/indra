package engine

import (
	"github.com/cybriq/qu"
	"go.uber.org/atomic"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/engine/ifc"
	"git-indra.lan/indra-labs/indra/pkg/engine/node"
	"git-indra.lan/indra-labs/indra/pkg/engine/responses"
	"git-indra.lan/indra-labs/indra/pkg/engine/sess"
	"git-indra.lan/indra-labs/indra/pkg/engine/sessions"
	"git-indra.lan/indra-labs/indra/pkg/engine/tpt"
	"git-indra.lan/indra-labs/indra/pkg/splice"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

type Engine struct {
	Responses *responses.Pending
	*sess.Manager
	*HiddenRouting
	*crypto.KeySet
	Load          atomic.Uint32
	TimeoutSignal qu.C
	Pause         qu.C
	ShuttingDown  atomic.Bool
	qu.C
}

type Params struct {
	tpt.Transport
	IDPrv           *crypto.Prv
	Node            *node.Node
	Nodes           []*node.Node
	NReturnSessions int
}

func NewEngine(p Params) (c *Engine, e error) {
	p.Node.Transport = p.Transport
	p.Node.Identity.Prv = p.IDPrv
	p.Node.Identity.Pub = crypto.DerivePub(p.IDPrv)
	var ks *crypto.KeySet
	if _, ks, e = crypto.NewSigner(); fails(e) {
		return
	}
	c = &Engine{
		Responses:     &responses.Pending{},
		KeySet:        ks,
		Manager:       sess.NewSessionManager(),
		HiddenRouting: NewHiddenrouting(),
		TimeoutSignal: qu.T(),
		Pause:         qu.T(),
		C:             qu.T(),
	}
	c.AddNodes(append([]*node.Node{p.Node}, p.Nodes...)...)
	// AddIntro a return session for receiving responses, ideally more of these
	// will be generated during operation and rotated out over time.
	for i := 0; i < p.NReturnSessions; i++ {
		c.AddSession(sessions.NewSessionData(nonce.NewID(), p.Node, 0, nil, nil, 5))
	}
	return
}

// Start a single thread of the Engine.
func (ng *Engine) Start() {
	log.T.Ln("starting engine")
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
	ng.Cleanup()
	ng.C.Q()
}

func (ng *Engine) HandleMessage(s *splice.Splice, pr ifc.Onion) {
	log.D.F("%s handling received message", ng.GetLocalNodeAddressString())
	s.SetCursor(0)
	s.Segments = s.Segments[:0]
	on := Recognise(s)
	if on != nil {
		log.D.Ln("magic", on.Magic())
		if fails(on.Decode(s)) {
			return
		}
		if pr != nil && on.Magic() != pr.Magic() {
			log.D.S(s.GetAll())
		}
		m := on.GetOnion()
		if m == nil {
			return
		}
		if fails(m.(ifc.Onion).Handle(s, pr, ng)) {
			log.W.S("unrecognised packet", s.GetAll().ToBytes())
		}
	}
}

func (ng *Engine) Handler() (out bool) {
	log.T.C(func() string {
		return ng.GetLocalNodeAddressString() + " awaiting message"
	})
	var prev ifc.Onion
	select {
	case <-ng.C.Wait():
		ng.Shutdown()
		out = true
		break
	case b := <-ng.ReceiveToLocalNode(0):
		s := splice.Load(b, slice.NewCursor())
		ng.HandleMessage(s, prev)
	case p := <-ng.GetLocalNode().Chan.Receive():
		log.D.F("incoming payment for %s: %v", p.ID, p.Amount)
		topUp := false
		ng.IterateSessions(func(s *sessions.Data) bool {
			if s.Preimage == p.Preimage {
				s.IncSats(p.Amount, false, "top-up")
				topUp = true
				log.T.F("topping up %x with %v", s.ID, p.Amount)
				return true
			}
			return false
		})
		if !topUp {
			ng.AddPendingPayment(p)
			log.T.F("awaiting session keys for preimage %s session ID %s",
				p.Preimage, p.ID)
		}
		// For now if we received this we return true. Later this will wait with
		// a timeout on the lnd node returning the success to trigger this.
		p.ConfirmChan <- true
	case <-ng.Pause:
		log.D.Ln("pausing", ng.GetLocalNodeAddressString())
		// For testing purposes we need to halt this Handler and discard channel
		// messages.
	out:
		for {
			select {
			case <-ng.GetLocalNode().Chan.Receive():
				log.D.Ln("discarding payments while in pause")
			case <-ng.ReceiveToLocalNode(0):
				log.D.Ln("discarding messages while in pause")
			case <-ng.C.Wait():
				break out
			case <-ng.Pause:
				// This will then resume to the top level select.
				log.D.Ln("unpausing", ng.GetLocalNodeAddressString())
				break out
			}
			
		}
	}
	return
}

var _ ifc.Ngin = &Engine{}

func (ng *Engine) GetLoad() byte               { return byte(ng.Load.Load()) }
func (ng *Engine) SetLoad(load byte)           { ng.Load.Store(uint32(load)) }
func (ng *Engine) Mgr() *sess.Manager          { return ng.Manager }
func (ng *Engine) Pending() *responses.Pending { return ng.Responses }
func (ng *Engine) Hidden() *HiddenRouting      { return ng.HiddenRouting }
func (ng *Engine) KillSwitch() qu.C            { return ng.C }
func (ng *Engine) Keyset() *crypto.KeySet      { return ng.KeySet }
