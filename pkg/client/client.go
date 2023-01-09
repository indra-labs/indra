package client

import (
	"math"
	"net/netip"
	"sync"
	"time"

	"github.com/cybriq/qu"
	"github.com/indra-labs/indra"
	"github.com/indra-labs/indra/pkg/ifc"
	"github.com/indra-labs/indra/pkg/key/cloak"
	"github.com/indra-labs/indra/pkg/key/prv"
	"github.com/indra-labs/indra/pkg/key/pub"
	"github.com/indra-labs/indra/pkg/key/signer"
	log2 "github.com/indra-labs/indra/pkg/log"
	"github.com/indra-labs/indra/pkg/node"
	"github.com/indra-labs/indra/pkg/nonce"
	"github.com/indra-labs/indra/pkg/session"
	"github.com/indra-labs/indra/pkg/slice"
	"github.com/indra-labs/indra/pkg/wire"
	"github.com/indra-labs/indra/pkg/wire/confirm"
	"github.com/indra-labs/indra/pkg/wire/layer"
	"github.com/indra-labs/indra/pkg/wire/response"
	"github.com/indra-labs/indra/pkg/wire/reverse"
	"go.uber.org/atomic"
)

const (
	DefaultDeadline  = 10 * time.Minute
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
	session.Sessions
	PendingSessions []nonce.ID
	*confirm.Confirms
	ExitHooks response.Hooks
	sync.Mutex
	*signer.KeySet
	atomic.Bool
	qu.C
}

func New(tpt ifc.Transport, hdrPrv *prv.Key, no *node.Node,
	nodes node.Nodes) (c *Client, e error) {

	no.Transport = tpt
	no.HeaderPrv = hdrPrv
	no.HeaderPub = pub.Derive(hdrPrv)
	var ks *signer.KeySet
	if _, ks, e = signer.New(); check(e) {
		return
	}
	c = &Client{
		Confirms: confirm.NewConfirms(),
		Node:     no,
		Nodes:    nodes,
		KeySet:   ks,
		C:        qu.T(),
	}
	c.Sessions = c.Sessions.Add(session.New(no.ID, math.MaxUint64, 0))
	return
}

// Start a single thread of the Client.
func (cl *Client) Start() {
out:
	for {
		if cl.runner() {
			break out
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

// FindCloaked searches the client identity key and the Sessions for a match.
func (cl *Client) FindCloaked(clk cloak.PubKey) (hdr *prv.Key, pld *prv.Key) {
	var b cloak.Blinder
	copy(b[:], clk[:cloak.BlindLen])
	hash := cloak.Cloak(b, cl.Node.HeaderBytes)
	if hash == clk {
		hdr = cl.Node.HeaderPrv
		// there is no payload key for the node, only in sessions.
		return
	}
	for i := range cl.Sessions {
		hash = cloak.Cloak(b, cl.Sessions[i].HeaderBytes)
		if hash == clk {
			hdr = cl.Sessions[i].HeaderPrv
			pld = cl.Sessions[i].PayloadPrv
			return
		}
	}
	return
}

func (cl *Client) SendKeys(nodeID nonce.ID,
	hook func(cf nonce.ID)) (conf nonce.ID, e error) {

	var hdrPrv, pldPrv *prv.Key
	if hdrPrv, e = prv.GenerateKey(); check(e) {
		return
	}
	hdrPub := pub.Derive(hdrPrv)
	if pldPrv, e = prv.GenerateKey(); check(e) {
		return
	}
	pldPub := pub.Derive(pldPrv)
	n := cl.Nodes.FindByID(nodeID)
	selected := cl.Nodes.Select(SimpleSelector, n, 4)
	var hop [5]*node.Node
	hop[0], hop[1], hop[2], hop[3], hop[4] =
		selected[0], selected[1], n, selected[2], selected[3]
	conf = nonce.NewID()
	os := wire.SendKeys(conf, hdrPrv, pldPrv, cl.Node, hop, cl.KeySet)
	cl.RegisterConfirmation(hook, os[len(os)-1].(*confirm.OnionSkin).ID)
	cl.Sessions.Add(&session.Session{
		ID:           n.ID,
		Remaining:    1 << 16,
		HeaderPub:    hdrPub,
		HeaderBytes:  hdrPub.ToBytes(),
		PayloadPub:   pldPub,
		PayloadBytes: pldPub.ToBytes(),
		HeaderPrv:    hdrPrv,
		PayloadPrv:   pldPrv,
		Deadline:     time.Now().Add(DefaultDeadline),
	})
	o := os.Assemble()
	b := wire.EncodeOnion(o)
	cl.Send(hop[0].AddrPort, b)
	return
}

func (cl *Client) Cleanup() {
	// Do cleanup stuff before shutdown.
}

func (cl *Client) Shutdown() {
	if cl.Bool.Load() {
		return
	}
	log.I.Ln("shutting down client", cl.Node.AddrPort.String())
	cl.Bool.Store(true)
	cl.C.Q()
}

func (cl *Client) Send(addr *netip.AddrPort, b slice.Bytes) {
	// first search if we already have the node available with connection
	// open.
	as := addr.String()
	for i := range cl.Nodes {
		if as == cl.Nodes[i].Addr {
			cl.Nodes[i].Transport.Send(b)
			return
		}
	}
	// If we got to here none of the addresses matched, and we need to
	// establish a new peer connection to them.

}
