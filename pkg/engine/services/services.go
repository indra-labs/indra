// Package services defines the base data structure for a service.
//
// This includes the port specification, the fee rate on the service, and the transport abstraction that opens a channel for messages to the service, or its listener depending on which side this structure is used.
package services

import "github.com/indra-labs/indra/pkg/engine/tpt"

type (
	Service struct {
		Port      uint16
		RelayRate uint32
		tpt.Transport
	}
	Services []*Service
)
