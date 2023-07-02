// Package ont defines interfaces for the engine: Ngin and Onion coding.Codec subtypes, and some helpers that use the abstraction.
package ont

import (
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/engine/coding"
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

// Ngin is the generic interface for onion encoders to access the engine without
// tying the dependencies together.
type Ngin interface {
	HandleMessage(s *splice.Splice, pr Onion)
	GetLoad() byte
	SetLoad(byte)
	Mgr() *sess.Manager
	Pending() *responses.Pending
	GetHidden() *hidden.Hidden
	WaitForShutdown() <-chan struct{}
	Keyset() *crypto.KeySet
}

// Onion are messages that can be layered over each other and have
// a set of processing instructions for the data in them, and, if relevant,
// how to account for them in sessions.
type Onion interface {
	coding.Codec
	Wrap(inner Onion)
	Handle(s *splice.Splice, p Onion, ni Ngin) (e error)
	Account(res *sess.Data, sm *sess.Manager, s *sessions.Data,
		last bool) (skip bool, sd *sessions.Data)
}

// Encode is the generic encoder for an onion, all onions can be encoded with it.
func Encode(d coding.Codec) (s *splice.Splice) {
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
