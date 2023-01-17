package service

import (
	"github.com/indra-labs/indra/pkg/types"
)

type Service struct {
	Port uint16
	types.Transport
}

type Services []*Service
