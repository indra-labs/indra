// Package confirmation provides an onion message type that simply returns a confirmation for an associated nonce.ID of a previous message that we want to confirm was received.
package confirmation

import (
	"github.com/indra-labs/indra/pkg/codec"
	"github.com/indra-labs/indra/pkg/codec/ont"
	"github.com/indra-labs/indra/pkg/codec/reg"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/engine/magic"
	"github.com/indra-labs/indra/pkg/engine/sess"
	"github.com/indra-labs/indra/pkg/engine/sessions"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/splice"
)

var (
	log   = log2.GetLogger()
	fails = log.E.Chk
)

const (
	Magic = "conf"
	Len   = magic.Len + nonce.IDLen
)

// Confirmation is simply a nonce that associates with a pending circuit
// transmission.
//
// If a reply is required there needs to be a RoutingHeader and cipher/nonce set.
type Confirmation struct {

	// ID of the request this response relates to.
	ID nonce.ID
}

// Account simply records the message ID, which will be recognised in the pending
// responses cache.
func (x *Confirmation) Account(res *sess.Data, sm *sess.Manager,
	s *sessions.Data, last bool) (skip bool, sd *sessions.Data) {
	res.ID = x.ID
	return
}

// Decode a splice.Splice's next bytes into a Confirmation.
func (x *Confirmation) Decode(s *splice.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), Len-magic.Len,
		Magic); fails(e) {
		return
	}
	s.ReadID(&x.ID)
	return
}

// Encode a Balance into a splice.Splice's next bytes.
func (x *Confirmation) Encode(s *splice.Splice) (e error) {
	s.Magic(Magic).ID(x.ID)
	return
}

// GetOnion returns nothing because there isn't an onion inside a Confirmation.
func (x *Confirmation) GetOnion() interface{} { return nil }

// Handle searches for a pending response and if it matches, runs the stored
// callbacks attached to it.
func (x *Confirmation) Handle(s *splice.Splice, p ont.Onion, ng ont.Ngin) (e error) {
	// When a confirmation arrives check if it is registered for and run the
	// hook that was registered with it.
	ng.Pending().ProcessAndDelete(x.ID, nil, s.GetAll())
	return
}

// Len returns the length of bytes required to encode the Confirmation.
func (x *Confirmation) Len() int { return Len }

// Magic bytes that identify this message
func (x *Confirmation) Magic() string { return Magic }

// Wrap is a no-op because a Confirmation is terminal.
func (x *Confirmation) Wrap(inner ont.Onion) {}

// New creates a new Confirmation.
func New(id nonce.ID) ont.Onion { return &Confirmation{ID: id} }

// Gen is a factory function to generate an Ad.
func Gen() codec.Codec { return &Confirmation{} }

func init() { reg.Register(Magic, Gen) }
