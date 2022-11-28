package transport

import (
	"github.com/Indra-Labs/indra/pkg/ifc"
)

type Dispatcher chan ifc.Message

func (d Dispatcher) Send(b ifc.Message) {
	d <- b
}

func (d Dispatcher) Receive() <-chan ifc.Message {
	return d
}

func NewDispatcher(bufs int) Dispatcher {
	return make(Dispatcher, bufs)
}
