package relay

import (
	"sync"

	"github.com/cybriq/qu"
	"go.uber.org/atomic"

	"git-indra.lan/indra-labs/indra"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/signer"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/traffic"
	"git-indra.lan/indra-labs/indra/pkg/types"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

type Engine struct {
	sync.Mutex
	*traffic.Node
	*PendingResponses
	*traffic.SessionManager
	*signer.KeySet
	Load         byte
	ShuttingDown atomic.Bool
	qu.C
}

func NewEngine(tpt types.Transport, hdrPrv *prv.Key, no *traffic.Node,
	nodes []*traffic.Node) (c *Engine, e error) {

	no.Transport = tpt
	no.IdentityPrv = hdrPrv
	no.IdentityPub = pub.Derive(hdrPrv)
	var ks *signer.KeySet
	if _, ks, e = signer.New(); check(e) {
		return
	}
	c = &Engine{
		Node:             no,
		PendingResponses: &PendingResponses{},
		KeySet:           ks,
		SessionManager:   traffic.NewSessionManager(),
		C:                qu.T(),
	}
	c.AddNodes(nodes...)
	// Add a return session for receiving responses, ideally more of these will
	// be generated during operation and rotated out over time.
	c.AddSession(traffic.NewSession(nonce.NewID(), no, 0, nil, nil, 5))
	return
}

// Start a single thread of the Engine.
func (eng *Engine) Start() {
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
	log.T.C(func() string {
		return "shutting down client " + eng.Node.AddrPort.String()
	})
	eng.ShuttingDown.Store(true)
	eng.C.Q()
}
