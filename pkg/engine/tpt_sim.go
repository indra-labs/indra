package engine

import (
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

type ByteChan chan slice.Bytes

func NewSim(bufs int) ByteChan { return make(ByteChan, bufs) }
func (s ByteChan) Send(b slice.Bytes) {
	s <- b
}
func (s ByteChan) Receive() <-chan slice.Bytes {
	return s
}
