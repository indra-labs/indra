package engine

import (
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

type Pipeline struct {
	in ByteChan
	Transport
}

func NewPipeline(bufs int) *Pipeline {
	return &Pipeline{in: make(ByteChan, bufs)}
}

func (p *Pipeline) Send(b slice.Bytes) {
	if p.Transport != nil {
		p.Transport.Send(b)
	} else {
		p.in <- b
	}
}
func (p *Pipeline) Receive() <-chan slice.Bytes {
	if p.Transport != nil {
		return p.Transport.Receive()
	} else {
		return p.in
	}
}
