package client

import (
	"net/netip"
	"sync"
	"time"

	"github.com/cybriq/qu"
	"github.com/davecgh/go-spew/spew"
	"github.com/indra-labs/indra"
	"github.com/indra-labs/indra/pkg/crypto/key/cloak"
	"github.com/indra-labs/indra/pkg/crypto/key/prv"
	"github.com/indra-labs/indra/pkg/crypto/key/pub"
	"github.com/indra-labs/indra/pkg/crypto/key/signer"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/node"
	"github.com/indra-labs/indra/pkg/onion/layers/confirm"
	"github.com/indra-labs/indra/pkg/onion/layers/crypt"
	"github.com/indra-labs/indra/pkg/onion/layers/response"
	"github.com/indra-labs/indra/pkg/onion/layers/reverse"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/traffic"
	"github.com/indra-labs/indra/pkg/types"
	"github.com/indra-labs/indra/pkg/util/slice"
	"go.uber.org/atomic"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

const (
	ReverseLayerLen  = reverse.Len + crypt.Len
	ReverseHeaderLen = 3 * ReverseLayerLen
)

type Client struct {
	*node.Node
	node.Nodes
	sync.Mutex
	*confirm.Confirms
	response.Hooks
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
	pld *prv.Key, sess *traffic.Session, identity bool) {

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
	cl.Node.IterateSessions(func(s *traffic.Session) (stop bool) {
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
		if as == cl.Nodes[i].AddrPort.String() {
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
