// Package ont defines interfaces for the engine: Ngin and Onion coding.Codec subtypes, and some helpers that use the abstraction.
package ont

import (
	"git.indra-labs.org/dev/ind/pkg/codec"
	"git.indra-labs.org/dev/ind/pkg/crypto"
	"git.indra-labs.org/dev/ind/pkg/engine/responses"
	"git.indra-labs.org/dev/ind/pkg/engine/sess"
	"git.indra-labs.org/dev/ind/pkg/engine/sessions"
	"git.indra-labs.org/dev/ind/pkg/hidden"
	log2 "git.indra-labs.org/dev/ind/pkg/proc/log"
	"git.indra-labs.org/dev/ind/pkg/util/splice"
)

var (
	log   = log2.GetLogger()
	fails = log.E.Chk
)

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
	codec.Codec

	// Wrap places another onion inside this onion's inner layer.
	Wrap(inner Onion)

	// Handle is the relay switching logic used by the Ngin on the Onion.
	Handle(s *splice.Splice, p Onion, ni Ngin) (e error)

	// Account sets up the bandwidth accounting for sending out an Onion.
	Account(res *sess.Data, sm *sess.Manager, s *sessions.Data,
		last bool) (skip bool, sd *sessions.Data)
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
