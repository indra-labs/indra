package transport

import "git-indra.lan/indra-labs/indra/pkg/util/slice"

// Transport is a generic interface for sending and receiving slices of bytes.
type Transport interface {
	Send(b slice.Bytes) (e error)
	Receive() <-chan slice.Bytes
}
