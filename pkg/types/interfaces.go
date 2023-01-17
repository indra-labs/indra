package types

import (
	"github.com/indra-labs/indra/pkg/util/slice"
)

type Transport interface {
	Send(b slice.Bytes)
	Receive() <-chan slice.Bytes
}
