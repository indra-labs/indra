// Package onion provides a set of functions to manage creating onion layered
// encryption for use with multi-hop Circuit protocol.
package onion

import (
	"fmt"

	"github.com/Indra-Labs/indra/pkg/node"
	"github.com/Indra-Labs/indra/pkg/nonce"
)

// A Circuit is a series of relays through which a message will be delivered.
// The Exit marks the index of the Hops slice that the message is relayed out of
// Indranet and the Hops after this index require the compound cipher and
// pre-made onion header that will be attached to the payload.
//
// The structure is not fixed in format to enable later creation of variations
// of longer and shorter Circuits and embedded in multi-path routes where
// packets are split laterally and delivered in parallel.
//
// Path diagnostic onions are encoded using circuits of 3 hops with the exit as
// the last.
type Circuit struct {
	ID   nonce.ID
	Hops node.Nodes
	Exit int
}

func (c *Circuit) GenerateOnion(exit []byte, messages [][]byte) (msg []byte,
	e error) {

	if len(messages) != len(c.Hops) {
		e = fmt.Errorf("mismatch of message count, %d messages, %d hops",
			len(messages), len(c.Hops))
		return
	}

	return
}

type Circuits []*Circuit

func New(id nonce.ID, hops node.Nodes, exit int) (c *Circuit) {
	c = &Circuit{id, hops, exit}
	return
}

func (cs Circuits) Find(id nonce.ID) (c *Circuit) {
	for i := range cs {
		if cs[i].ID == id {
			return cs[i]
		}
	}
	return
}

func (cs Circuits) Add(c *Circuit) (co Circuits) {
	co = append(cs, c)
	return
}

func (cs Circuits) Delete(id nonce.ID) (co Circuits, e error) {
	e = fmt.Errorf("circuit ID %x not found", id)
	for i := range cs {
		if cs[i].ID == id {
			return append(cs[:i], cs[i+1:]...), nil
		}
	}
	return
}
