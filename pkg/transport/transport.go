package transport

import (
	"github.com/Indra-Labs/indra/pkg/ifc"
)

type Dispatcher chan ifc.Bytes

func (d Dispatcher) Send(b ifc.Bytes) {
	d <- b
}

func (d Dispatcher) Receive() <-chan ifc.Bytes {
	return d
}

func NewDispatcher(bufs int) Dispatcher {
	return make(Dispatcher, bufs)
}
