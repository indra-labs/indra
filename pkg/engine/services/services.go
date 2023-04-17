package services

import "git-indra.lan/indra-labs/indra/pkg/engine/transport"

type Service struct {
	Port      uint16
	RelayRate int
	transport.Transport
}

type Services []*Service
