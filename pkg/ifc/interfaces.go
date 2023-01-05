package ifc

import (
	"github.com/Indra-Labs/indra/pkg/slice"
)

// var (
// 	log   = log2.GetLogger(indra.PathBase)
// 	check = log.E.Chk
// )

type Transport interface {
	Send(b slice.Bytes)
	Receive() <-chan slice.Bytes
}
