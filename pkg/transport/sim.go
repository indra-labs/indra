package transport

import (
	"git-indra.lan/indra-labs/indra"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

type Sim chan slice.Bytes

func NewSim(bufs int) Sim { return make(Sim, bufs) }
func (d Sim) Send(b slice.Bytes) {
	d <- b
}
func (d Sim) Receive() <-chan slice.Bytes {
	return d
}
