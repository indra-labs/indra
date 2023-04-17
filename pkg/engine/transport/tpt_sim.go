package transport

import (
	"context"
	
	"git-indra.lan/indra-labs/indra/pkg/engine/tpt"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

type ByteChan chan slice.Bytes

func NewByteChan(bufs int) ByteChan { return make(ByteChan, bufs) }

func (s ByteChan) Send(b slice.Bytes) error {
	s <- b
	return nil
}

func (s ByteChan) Receive() <-chan slice.Bytes {
	return s
}

// DuplexByteChan is intended to be connected up in chains with other processing
// steps as a pipeline. The send and receive functions send bytes to their
// respective send and receive channels, and the processing is added by a
// consuming type by listening to the send channel for requests to send, and
// forwarding data from the upstream to the receive channel.
type DuplexByteChan struct {
	Receiver, Sender tpt.Transport
}

func NewDuplexByteChan(bufs int) *DuplexByteChan {
	return &DuplexByteChan{NewByteChan(bufs), NewByteChan(bufs)}
}

// NewSimDuplex creates a DuplexByteChan that behaves like a single ByteChan by
// forwarding from the send channel to the receive channel. This creates
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

func (d *DuplexByteChan) Send(b slice.Bytes) (e error) {
	d.Sender.Send(b)
	return
}

func (d *DuplexByteChan) Receive() (C <-chan slice.Bytes) {
	return d.Receiver.Receive()
}
