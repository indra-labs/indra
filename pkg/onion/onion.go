// Package onion provides a set of functions to manage creating onion layered
// encryption for use with multi-hop Circuit protocol.
package onion

import (
	"github.com/Indra-Labs/indra/pkg/node"
	"github.com/Indra-Labs/indra/pkg/nonce"
)

// A Return is a return path for a hop. These are optional to a Circuit and used
// when a path times out in order to diagnose which hop is failing.
type Return struct {
	Hops [3]*node.Node
}

type Returns []*Return

// A Circuit is a series of relays through which a message will be delivered.
// The Exit marks the index of the Hops slice that the message is relayed out of
// Indranet and the Hops after this index require the compound cipher and
// pre-made onion header that will be attached to the payload.
//
// Trace is a collection of Returns that corresponds each of the Hops, this is
// optional and used for path liveness diagnostics when the Circuit times out.
//
// The structure is not fixed in format to enable later creation of variations
// of longer and shorter Circuits and embedded in multi-path routes where
// packets are split laterally and delivered in parallel.
type Circuit struct {
	ID    nonce.ID
	Hops  node.Nodes
	Exit  int
	Trace Returns
}

type Circuits []*Circuit

func New(id nonce.ID, hops node.Nodes, exit int, trace Returns) (c *Circuit) {
	c = &Circuit{id, hops, exit, trace}
	return
}
