package services

import "github.com/indra-labs/indra/pkg/engine/tpt"

type (
	Service struct {
		Port      uint16
		RelayRate int
		tpt.Transport
	}
	Services []*Service
)
