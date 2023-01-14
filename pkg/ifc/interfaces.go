package ifc

import (
	"github.com/indra-labs/indra/pkg/slice"
)

type Transport interface {
	Send(b slice.Bytes)
	Receive() <-chan slice.Bytes
}
