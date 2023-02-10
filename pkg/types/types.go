package types

import (
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

// Onion is an interface for the layers of messages each encrypted inside a
// OnionSkin, which provides the cipher for the inner layers inside it.
type Onion interface {
	Encode(b slice.Bytes, c *slice.Cursor)
	Decode(b slice.Bytes, c *slice.Cursor) (e error)
	Len() int
	Inner() Onion
	Insert(on Onion)
}

type Transport interface {
	Send(b slice.Bytes)
	Receive() <-chan slice.Bytes
}
