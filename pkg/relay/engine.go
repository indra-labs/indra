package relay

import (
	"github.com/cybriq/qu"
	"go.uber.org/atomic"
	
	"git-indra.lan/indra-labs/indra"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/signer"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/relay/types"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
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

type EngineParams struct {
	Tpt             types.Transport
	IDPrv           *prv.Key
	No              *Node
	Nodes           []*Node
	NReturnSessions int
}

func NewEngine(p EngineParams) (c *Engine, e error) {
	p.No.Transport = p.Tpt
	p.No.IdentityPrv = p.IDPrv
	p.No.IdentityPub = pub.Derive(p.IDPrv)
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
	c.AddNodes(append([]*Node{p.No}, p.Nodes...)...)
	// AddIntro a return session for receiving responses, ideally more of these will
	// be generated during operation and rotated out over time.
	for i := 0; i < p.NReturnSessions; i++ {
		c.AddSession(NewSession(nonce.NewID(), p.No, 0, nil, nil, 5))
	}
	return
}

// Start a single thread of the Engine.
func (eng *Engine) Start() {
	log.D.Ln("starting engine")
	for {
		if eng.handler() {
			break
		}
	}
}

// Cleanup closes and flushes any resources the client opened that require sync
// in order to reopen correctly.
func (eng *Engine) Cleanup() {
	// Do cleanup stuff before shutdown.
}

// Shutdown triggers the shutdown of the client and the Cleanup before
// finishing.
func (eng *Engine) Shutdown() {
	if eng.ShuttingDown.Load() {
		return
	}
	eng.ShuttingDown.Store(true)
	log.T.C(func() string {
		return "shutting down client " + eng.GetLocalNodeAddress().String()
	})
	eng.Cleanup()
	eng.C.Q()
}
