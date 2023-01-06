package node

import (
	"github.com/indra-labs/indra/pkg/ifc"
)

type Service struct {
	Port uint16
	ifc.Transport
}

type Services []*Service
