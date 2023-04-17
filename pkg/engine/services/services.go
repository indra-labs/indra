package services

import "git-indra.lan/indra-labs/indra/pkg/engine/tpt"

type Service struct {
	Port      uint16
	RelayRate int
	tpt.Transport
}

type Services []*Service
