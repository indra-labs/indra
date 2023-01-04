package transport

import (
	"github.com/Indra-Labs/indra"
	log2 "github.com/Indra-Labs/indra/pkg/proc/pkg/log"
	"github.com/Indra-Labs/indra/pkg/slice"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

type Sim chan slice.Bytes

func NewSim(bufs int) Sim { return make(Sim, bufs) }
func (d Sim) Send(b slice.Bytes) {
	// log.I.S("sending", b.ToBytes())
	// log.I.S(runtime.Caller(1))
	d <- b
}
func (d Sim) Receive() <-chan slice.Bytes {
	// log.I.Ln("receiving")
	return d
}
