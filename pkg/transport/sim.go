package transport

import (
	"github.com/Indra-Labs/indra/pkg/ifc"
)

type Sim chan ifc.Message

func (d Sim) Send(b ifc.Message) {
	d <- b
}

func (d Sim) Receive() <-chan ifc.Message {
	return d
}
