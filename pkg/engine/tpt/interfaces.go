// Package tpt provides the definition of the interface Transport, which is an abstraction used for reading and writing to peers via transport.Transport.
package tpt

import "github.com/indra-labs/indra/pkg/util/slice"

// Transport is a generic interface for sending and receiving slices of bytes.
type Transport interface {

	// Send delivers a byte buffer along the Transport.
	Send(b slice.Bytes) (e error)

	// Receive waits for incoming messages and delivers the byte buffer to the caller
	// via a channel.
	Receive() <-chan slice.Bytes
}
