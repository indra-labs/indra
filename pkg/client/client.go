package client

import (
	"net/netip"
	"sync"
	"time"

	"github.com/cybriq/qu"
	"github.com/davecgh/go-spew/spew"
	"github.com/indra-labs/indra"
	"github.com/indra-labs/indra/pkg/ifc"
	"github.com/indra-labs/indra/pkg/key/cloak"
	"github.com/indra-labs/indra/pkg/key/prv"
	"github.com/indra-labs/indra/pkg/key/pub"
	"github.com/indra-labs/indra/pkg/key/signer"
	log2 "github.com/indra-labs/indra/pkg/log"
	"github.com/indra-labs/indra/pkg/node"
	"github.com/indra-labs/indra/pkg/nonce"
	"github.com/indra-labs/indra/pkg/slice"
	"github.com/indra-labs/indra/pkg/wire/confirm"
	"github.com/indra-labs/indra/pkg/wire/layer"
	"github.com/indra-labs/indra/pkg/wire/response"
	"github.com/indra-labs/indra/pkg/wire/reverse"
	"go.uber.org/atomic"
)

const (
	ReverseLayerLen  = reverse.Len + layer.Len
	ReverseHeaderLen = 3 * ReverseLayerLen
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

type Client struct {
	*node.Node
	node.Nodes
	*confirm.Confirms
	response.Hooks
	sync.Mutex
	*signer.KeySet
	ShuttingDown atomic.Bool
	qu.C
}

func New(tpt ifc.Transport, hdrPrv *prv.Key, no *node.Node,
	nodes node.Nodes) (c *Client, e error) {

	no.Transport = tpt
	no.IdentityPrv = hdrPrv
	no.IdentityPub = pub.Derive(hdrPrv)
	var ks *signer.KeySet
	if _, ks, e = signer.New(); check(e) {
		return
	}
	// Add our first return session.
	no.AddSession(node.NewSession(nonce.NewID(), no, 0, nil, nil, 5))
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
		if cl.runner() {
			break
		}
	}
}

func (cl *Client) RegisterConfirmation(hook confirm.Hook,
	cnf nonce.ID) {

	cl.Confirms.Add(&confirm.Callback{
		ID:   cnf,
		Time: time.Now(),
		Hook: hook,
	})
}

// FindCloaked searches the client identity key and the sessions for a match. It
// returns the session as well, though not all users of this function will need
// this.
func (cl *Client) FindCloaked(clk cloak.PubKey) (hdr *prv.Key,
	pld *prv.Key, sess *node.Session, identity bool) {

	var b cloak.Blinder
	copy(b[:], clk[:cloak.BlindLen])
	hash := cloak.Cloak(b, cl.Node.IdentityBytes)
	if hash == clk {
		log.T.Ln("encrypted to identity key")
		hdr = cl.Node.IdentityPrv
		// there is no payload key for the node, only in sessions.
		identity = true
		return
	}
	var i int
	cl.Node.IterateSessions(func(s *node.Session) (stop bool) {
		hash = cloak.Cloak(b, s.HeaderBytes)
		if hash == clk {
			log.T.F("found cloaked key in session %d", i)
			hdr = s.HeaderPrv
			pld = s.PayloadPrv
			sess = s
			return true
		}
		i++
		return
	})
	return
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

// Send a message to a peer via their AddrPort.
func (cl *Client) Send(addr *netip.AddrPort, b slice.Bytes) {
	// first search if we already have the node available with connection
	// open.
	as := addr.String()
	for i := range cl.Nodes {
		if as == cl.Nodes[i].Addr {
			log.T.C(func() string {
				return cl.AddrPort.String() +
					" sending to " +
					addr.String() +
					"\n" +
					spew.Sdump(b.ToBytes())
			})
			cl.Nodes[i].Transport.Send(b)
			return
		}
	}
	// If we got to here none of the addresses matched, and we need to
	// establish a new peer connection to them, if we know of them (this
	// would usually be the reason this happens).

}
