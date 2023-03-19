package ngin

import (
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

type Sim chan slice.Bytes

func NewSim(bufs int) Sim { return make(Sim, bufs) }
func (d Sim) Send(b slice.Bytes) {
	d <- b
}
func (d Sim) Receive() <-chan slice.Bytes {
	return d
}

func (d Sim) Chain(t Transport) Transport {
	// TODO implement me
	panic("implement me")
}
