package transport

import (
	"github.com/Indra-Labs/indra/pkg/slice"
)

type Sim chan slice.Bytes

func NewSim(bufs int) Sim                 { return make(Sim, bufs) }
func (d Sim) Send(b slice.Bytes)          { d <- b }
func (d Sim) Receive() <-chan slice.Bytes { return d }
