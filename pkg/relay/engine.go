package relay

import (
	"sync"
	"time"

	"github.com/cybriq/qu"
	"go.uber.org/atomic"

	"git-indra.lan/indra-labs/indra"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/signer"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/node"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/confirm"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/traffic"
	"git-indra.lan/indra-labs/indra/pkg/types"
)

var (
	log   = log2.GetLogger(indrabase.PathBase)
	check = log.E.Chk
)

type Engine struct {
	*node.Node
	node.Nodes
	sync.Mutex
	Load byte
	*confirm.Confirms
	Pending PendingResponses
	*signer.KeySet
	ShuttingDown atomic.Bool
	qu.C
}

func NewEngine(tpt types.Transport, hdrPrv *prv.Key, no *node.Node,
	nodes node.Nodes, timeout time.Duration) (c *Engine, e error) {

	no.Transport = tpt
	no.IdentityPrv = hdrPrv
	no.IdentityPub = pub.Derive(hdrPrv)
	var ks *signer.KeySet
	if _, ks, e = signer.New(); check(e) {
		return
	}
	// Add our first return session.
	no.AddSession(traffic.NewSession(nonce.NewID(), no.Peer, 0, nil, nil, 5))
	c = &Engine{
		Confirms: confirm.NewConfirms(),
		Node:     no,
		Nodes:    nodes,
		KeySet:   ks,
		C:        qu.T(),
	}
	c.Pending.Timeout = timeout
	return
}

// Start a single thread of the Engine.
func (en *Engine) Start() {
	for {
		if en.handler() {
			break
		}
	}
}

func (en *Engine) RegisterConfirmation(hook confirm.Hook,
	cnf nonce.ID) {

	if hook == nil {
		return
	}
	en.Confirms.Add(&confirm.Callback{
		ID:   cnf,
		Time: time.Now(),
		Hook: hook,
	})
}

// Cleanup closes and flushes any resources the client opened that require sync
// in order to reopen correctly.
func (en *Engine) Cleanup() {
	// Do cleanup stuff before shutdown.
}

// Shutdown triggers the shutdown of the client and the Cleanup before
// finishing.
func (en *Engine) Shutdown() {
	if en.ShuttingDown.Load() {
		return
	}
	log.T.C(func() string {
		return "shutting down client " + en.Node.AddrPort.String()
	})
	en.ShuttingDown.Store(true)
	en.C.Q()
}
