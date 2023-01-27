package service

import (
	"github.com/indra-labs/indra/pkg/types"
	"github.com/indra-labs/lnd/lnd/lnwire"
)

type Service struct {
	Port      uint16
	RelayRate lnwire.MilliSatoshi
	types.Transport
}

type Services []*Service
