// Package services defines the base data structure for a service.
//
// This includes the port specification, the fee rate on the service, and the transport abstraction that opens a channel for messages to the service, or its listener depending on which side this structure is used.
package services

import "git.indra-labs.org/dev/ind/pkg/engine/tpt"

type (
	// Service is a specification for a publicly accessible service available at a
	// relay.
	//
	// Through this mechanism relay operators can effectively create a paywall for a
	// service, or at least cover their operating costs. Hidden services can do this
	// also, with server side anonymity.
	//
	// todo: hidden services need a session type.
	Service struct {

		// Port specifies the type of service based on the well known port used by the
		// protocol. For bitcoin, for example, this would be 8333 for its peer to peer
		// listener, for SSH it would be 22, and so on.
		Port uint16

		// RelayRate is the fee in mSAT for megabytes forwarded to and returned from the
		// service.
		RelayRate uint32

		// Transport is a channel that will have a network handler at the other end to
		// dispatch and return replies to the Engine.
		tpt.Transport
	}

	// Services is a collection of services.
	Services []*Service
)
