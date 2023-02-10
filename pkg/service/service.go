package service

import (
	"git-indra.lan/indra-labs/lnd/lnd/lnwire"

	"git-indra.lan/indra-labs/indra/pkg/types"
)

type Service struct {
	Port      uint16
	RelayRate lnwire.MilliSatoshi
	types.Transport
}

type Services []*Service
