package node

import (
	"github.com/Indra-Labs/indra/pkg/ifc"
)

type Service struct {
	Port uint16
	ifc.Transport
}

type Services []*Service
