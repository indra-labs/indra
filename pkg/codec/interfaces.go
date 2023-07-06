// Package codec defines an interface for encoding and decoding message packets in the Indra network.
//
// These are implemented for onion messages, advertisements and other peer messages that are not for relaying.
package codec

import (
	"github.com/indra-labs/indra/pkg/util/splice"
)

// Codec is a unit of data that can be read and written from a binary form. All
// Onion are Codec but not all Codec are Onion. Codec is also used for the
// Dispatcher's message headers.
type Codec interface {
	Magic() string
	Encode(s *splice.Splice) (e error)
	Decode(s *splice.Splice) (e error)
	Len() int
	Unwrap() interface{}
}
