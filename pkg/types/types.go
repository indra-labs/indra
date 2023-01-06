package types

import (
	"github.com/indra-labs/indra/pkg/slice"
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
