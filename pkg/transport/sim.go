package transport

import (
	"github.com/Indra-Labs/indra/pkg/slice"
)

type Sim chan slice.Bytes

func (d Sim) Send(b slice.Bytes) {
	d <- b
}

func (d Sim) Receive() <-chan slice.Bytes {
	return d
}

func NewSim(bufs int) Dispatcher {
	return make(Dispatcher, bufs)
}
