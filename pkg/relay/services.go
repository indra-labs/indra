package relay

import "git-indra.lan/indra-labs/indra/pkg/types"

type Service struct {
	Port      uint16
	RelayRate int
	types.Transport
}

type Services []*Service
