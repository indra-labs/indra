package relay

import (
	"time"
	
	"github.com/cybriq/qu"
	"go.uber.org/atomic"
	
	"git-indra.lan/indra-labs/indra"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/signer"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/types"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

const DefaultTimeout = time.Second

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

func NewEngine(tpt types.Transport, idPrv *prv.Key, no *Node,
	nodes []*Node, nReturnSessions int) (c *Engine, e error) {
	
	no.Transport = tpt
	no.IdentityPrv = idPrv
	no.IdentityPub = pub.Derive(idPrv)
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
	c.AddNodes(append([]*Node{no}, nodes...)...)
	// AddIntro a return session for receiving responses, ideally more of these will
	// be generated during operation and rotated out over time.
	for i := 0; i < nReturnSessions; i++ {
		c.AddSession(NewSession(nonce.NewID(), no, 0, nil, nil, 5))
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
