package client

import (
	"fmt"

	"github.com/indra-labs/indra/pkg/node"
	"github.com/indra-labs/indra/pkg/nonce"
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

// EncodeOnion uses a Circuit to create an onion message. An onion message
// consists of a series of layers containing the IP address to forward the
// attached payload and at the Exit hop there is an arbitrary blob of data.
//
// When there is additional hops after the Exit, these hops are interpreted to
// mean to expect a response to pass from the Exit hop and these layers contain
// a compound cipher for the remainder of the path.
func (c *Circuit) EncodeOnion(message []byte) (msg []byte, e error) {

	return
}

type Circuits []*Circuit

func NewCircuit(id nonce.ID, hops node.Nodes, exit int) (c *Circuit) {
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
