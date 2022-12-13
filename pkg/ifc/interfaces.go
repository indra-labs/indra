package ifc

import (
	"github.com/Indra-Labs/indra"
	"github.com/Indra-Labs/indra/pkg/slice"
	log2 "github.com/cybriq/proc/pkg/log"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

type Transport interface {
	Send(b slice.Bytes)
	Receive() <-chan slice.Bytes
}
