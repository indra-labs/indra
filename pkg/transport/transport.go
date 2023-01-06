package transport

import (
	"github.com/indra-labs/indra/pkg/slice"
)

type Dispatcher chan slice.Bytes

func (d Dispatcher) Send(b slice.Bytes) {
	d <- b
}

func (d Dispatcher) Receive() <-chan slice.Bytes {
	return d
}

func NewDispatcher(bufs int) Dispatcher {
	return make(Dispatcher, bufs)
}
