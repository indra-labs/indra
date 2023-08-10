package transport

import (
	"context"
	"git.indra-labs.org/dev/ind/pkg/engine/tpt"
	"git.indra-labs.org/dev/ind/pkg/util/slice"
)

type (
	// ByteChan is the most primitive form of an atomic FIFO queue used to
	// dispatch jobs to a network I/O handler.
	ByteChan chan slice.Bytes

	// DuplexByteChan is intended to be connected up in chains with other
	// processing steps as a pipeline. The send and receive functions send bytes
	// to their respective send and receive channels, and the processing is
	// added by a consuming type by listening to the send channel for requests
	// to send, and forwarding data from the upstream to the recieve channel.
	DuplexByteChan struct {
		// Receiver and Sender can send and receive in parallel.
		Receiver, Sender tpt.Transport
	}
)

// Receive messages from the receiver channel of the DuplexByteChan.
func (d *DuplexByteChan) Receive() (C <-chan slice.Bytes) {
	return d.Receiver.Receive()
}

// Send messages to the sender channel of the DuplexByteChan.
func (d *DuplexByteChan) Send(b slice.Bytes) (e error) {
	d.Sender.Send(b)
	return
}

// NewByteChan creates a new ByteChan with a specified number of buffers.
func NewByteChan(bufs int) ByteChan { return make(ByteChan, bufs) }

// NewDuplexByteChan creates a new DuplexByteChan with each of the two channels
// given a specified number of buffer queue slots.
func NewDuplexByteChan(bufs int) *DuplexByteChan {
	return &DuplexByteChan{Receiver: NewByteChan(bufs), Sender: NewByteChan(0)}
}

// NewSimDuplex creates a DuplexByteChan that behaves like a single ByteChan by
// forwarding from the send channel to the receiver channel. This creates
// something like a virtual in memory packet connection, as used in many of the
// Onion tests for testing correct forwarding without a full network.
//
// A network-using version of the same tests should also work exactly the same.
func NewSimDuplex(bufs int, ctx context.Context) (d *DuplexByteChan) {
	d = NewDuplexByteChan(bufs)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case b := <-d.Sender.Receive():
				d.Receiver.Send(b)
			}
		}
	}()
	return
}

// Receive returns the receiving side of the simplex ByteChan.
func (s ByteChan) Receive() <-chan slice.Bytes {
	return s
}

// Send the provided buffer to the ByteChan.
func (s ByteChan) Send(b slice.Bytes) error {
	s <- b
	return nil
}
