package client

import (
	"time"

	"github.com/indra-labs/indra/pkg/onion/layers/delay"
	"github.com/indra-labs/indra/pkg/types"
	"github.com/indra-labs/indra/pkg/util/slice"
)

func (cl *Client) delay(on *delay.Layer, b slice.Bytes,
	c *slice.Cursor, prev types.Onion) {

	// this is a message to hold the message in the buffer until a duration
	// elapses. The accounting for the remainder of the message adds a
	// factor to the effective byte consumption in accordance with the time
	// to be stored.
	// todo: accounting
	select {
	case <-time.After(on.Duration):
	}
	cl.handleMessage(BudgeUp(b, *c), on)
}
