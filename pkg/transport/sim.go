package transport

import (
	"github.com/Indra-Labs/indra/pkg/ifc"
)

type Sim chan ifc.Bytes

func (d Sim) Send(b ifc.Bytes) {
	d <- b
}

func (d Sim) Receive() <-chan ifc.Bytes {
	return d
}

func NewSim(bufs int) Dispatcher {
	return make(Dispatcher, bufs)
}
