// Package codec defines an interface for encoding and decoding message packets in the Indra network.
//
// These are implemented for onion messages, advertisements and other peer messages that are not for relaying.
package codec

import (
	log2 "git.indra-labs.org/dev/ind/pkg/proc/log"
	"git.indra-labs.org/dev/ind/pkg/util/splice"
)

var (
	log   = log2.GetLogger()
	fails = log.E.Chk
)

// Codec is a unit of data that can be read and written from a binary form. All
// Onion are Codec but not all Codec are Onion. Codec is also used for the
// Dispatcher's message headers.
type Codec interface {

	// Magic is a 4 byte string identifying the type of the following message bytes.
	Magic() string

	// Encode uses the Codec's contents to encode into the splice.Splice next bytes.
	Encode(s *splice.Splice) (e error)

	// Decode reads in the data in the next bytes of the splice.Splice to populate this Codec.
	Decode(s *splice.Splice) (e error)

	// Len returns the number of bytes required to encode this Codec message (including Magic).
	//
	// This function must panic if called on a nil pointer as unconfigured
	// messages cannot yield a valid length value in many cases.
	Len() int

	// Unwrap gives access to any further layers embedded inside this (specifically, the Onion inside).
	Unwrap() interface{}
}

func MustNotBeNil(c Codec) {
	if c == nil {
		panic("cannot compute length without struct fields")
	}
}

// Encode is the generic encoder for a Codec, all can be encoded with it.
func Encode(d Codec) (s *splice.Splice) {
	s = splice.New(d.Len())
	fails(d.Encode(s))
	return
}
