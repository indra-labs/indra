package client

import (
	"sync"
	"time"

	"github.com/cybriq/qu"
	"github.com/indra-labs/indra"
	"github.com/indra-labs/indra/pkg/crypto/key/prv"
	"github.com/indra-labs/indra/pkg/crypto/key/pub"
	"github.com/indra-labs/indra/pkg/crypto/key/signer"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/node"
	"github.com/indra-labs/indra/pkg/onion/layers/confirm"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/traffic"
	"github.com/indra-labs/indra/pkg/types"
	"go.uber.org/atomic"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

type Client struct {
	*node.Node
	node.Nodes
	sync.Mutex
	Load byte
	*confirm.Confirms
	PendingResponses
	*signer.KeySet
	ShuttingDown atomic.Bool
	qu.C
}

func NewClient(tpt types.Transport, hdrPrv *prv.Key, no *node.Node,
	nodes node.Nodes) (c *Client, e error) {

	no.Transport = tpt
	no.IdentityPrv = hdrPrv
	no.IdentityPub = pub.Derive(hdrPrv)
	var ks *signer.KeySet
	if _, ks, e = signer.New(); check(e) {
		return
	}
	// Add our first return session.
	no.AddSession(traffic.NewSession(nonce.NewID(), no.Peer, 0, nil, nil, 5))
	c = &Client{
		Confirms: confirm.NewConfirms(),
		Node:     no,
		Nodes:    nodes,
		KeySet:   ks,
		C:        qu.T(),
	}
	return
}

// Start a single thread of the Client.
func (cl *Client) Start() {
	for {
		if cl.handler() {
			break
		}
	}
}

func (cl *Client) RegisterConfirmation(hook confirm.Hook,
	cnf nonce.ID) {

	if hook == nil {
		return
	}
	cl.Confirms.Add(&confirm.Callback{
		ID:   cnf,
		Time: time.Now(),
		Hook: hook,
	})
}

// Cleanup closes and flushes any resources the client opened that require sync
// in order to reopen correctly.
func (cl *Client) Cleanup() {
	// Do cleanup stuff before shutdown.
}

// Shutdown triggers the shutdown of the client and the Cleanup before
// finishing.
func (cl *Client) Shutdown() {
	if cl.ShuttingDown.Load() {
		return
	}
	log.T.C(func() string {
		return "shutting down client " + cl.Node.AddrPort.String()
	})
	cl.ShuttingDown.Store(true)
	cl.C.Q()
}
