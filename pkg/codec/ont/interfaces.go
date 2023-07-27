// Package ont defines interfaces for the engine: Ngin and Onion coding.Codec subtypes, and some helpers that use the abstraction.
package ont

import (
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/engine/responses"
	"github.com/indra-labs/indra/pkg/engine/sess"
	"github.com/indra-labs/indra/pkg/engine/sessions"
	"github.com/indra-labs/indra/pkg/hidden"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/splice"
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
	Len() int

	// Unwrap gives access to any further layers embedded inside this (specifically, the Onion inside).
	Unwrap() interface{}
}

// Ngin is the generic interface for onion encoders to access the engine without
// tying the dependencies together.
type Ngin interface {

	// HandleMessage sets an engine to process an Onion.
	HandleMessage(s *splice.Splice, pr Onion)

	// GetLoad returns the current engine load level.
	GetLoad() byte

	// SetLoad sets the current engine load level.
	SetLoad(byte)

	// Mgr returns the pointer to the Session Manager of this Ngin.
	Mgr() *sess.Manager

	// Pending returns the pending responses handler.
	Pending() *responses.Pending

	// GetHidden returns the hidden services manager.
	GetHidden() *hidden.Hidden

	// WaitForShutdown returns a signal channel that returns after the shutdown
	// breaker is triggered.
	WaitForShutdown() <-chan struct{}

	// Keyset returns the scalar addition fast private key generator in use by the Ngin.
	Keyset() *crypto.KeySet
}

// Onion are messages that can be layered over each other and have
// a set of processing instructions for the data in them, and, if relevant,
// how to account for them in sessions.
type Onion interface {
	Codec

	// Wrap places another onion inside this onion's inner layer.
	Wrap(inner Onion)

	// Handle is the relay switching logic used by the Ngin on the Onion.
	Handle(s *splice.Splice, p Onion, ni Ngin) (e error)

	// Account sets up the bandwidth accounting for sending out an Onion.
	Account(res *sess.Data, sm *sess.Manager, s *sessions.Data,
		last bool) (skip bool, sd *sessions.Data)
}

// Encode is the generic encoder for a Codec, all can be encoded with it.
func Encode(d Codec) (s *splice.Splice) {
	s = splice.New(d.Len())
	fails(d.Encode(s))
	return
}

// Assemble takes a slice and inserts the tail into the onion of the head until
// there is no tail left.
func Assemble(o []Onion) (on Onion) {
	// First item is the outer crypt.
	on = o[0]
	// Iterate through the remaining layers.
	for _, oc := range o[1:] {
		on.Wrap(oc)
		// Next step we are inserting inside the one we just inserted.
		on = oc
	}
	// At the end, the first element contains references to every element
	// inside it.
	return o[0]
}
