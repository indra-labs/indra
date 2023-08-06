// Package whisper provides a message type for sending a message to a hidden service, or back to a hidden service client.
//
// These messages are the same for both sides after a route message forwards via the intro routing header an introducer receives for a hidden service, as there is no intermediary bridge like in rendezvous routing.
package whisper

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/indra-labs/indra/pkg/codec"
	"github.com/indra-labs/indra/pkg/codec/ont"
	"github.com/indra-labs/indra/pkg/codec/reg"
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/ciph"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/crypto/sha256"
	"github.com/indra-labs/indra/pkg/engine/consts"
	"github.com/indra-labs/indra/pkg/engine/magic"
	"github.com/indra-labs/indra/pkg/engine/sess"
	"github.com/indra-labs/indra/pkg/engine/sessions"
	"github.com/indra-labs/indra/pkg/hidden"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/slice"
	"github.com/indra-labs/indra/pkg/util/splice"
	"reflect"
)

var (
	log   = log2.GetLogger()
	fails = log.E.Chk
)

const (
	MessageMagic = "whis"

	// ReplyCiphersLen is
	//
	// Deprecated: this is now a variable length structure, Reverse is being
	// obsoleted in favour of Offset.
	ReplyCiphersLen = 2*consts.RoutingHeaderLen +
		6*sha256.Len +
		6*nonce.IVLen

	// MessageLen is
	//
	// Deprecated: this is now a variable length structure, Reverse is being
	// obsoleted in favour of Offset.
	MessageLen = magic.Len +
		2*nonce.IDLen +
		2*consts.RoutingHeaderLen +
		ReplyCiphersLen
)

// Message is the generic, peer to peer, bidirectional messade type for between a
// client and a hidden service.
//
// The message format is the same for both sides, as the connection is maintained
// by each side forwarding a new return path each message they send.
type Message struct {
	Forwards        [2]*sessions.Data
	Address         *crypto.Pub
	ID, Re          nonce.ID
	Forward, Return *hidden.ReplyHeader
	Payload         slice.Bytes
}

// NewMessage ... TODO
func NewMessage() (msg *Message) {

	return
}

// Account for the Message. The client obviously doesn't do anything with this.
//
// todo: how does hidden service bill? We need to establish hidden service session type.
func (x *Message) Account(res *sess.Data, sm *sess.Manager, s *sessions.Data, last bool) (skip bool, sd *sessions.Data) {
	res.ID = x.ID
	res.Billable = append(res.Billable, s.Header.Bytes)
	skip = true
	return
}

// Decode a Message from a provided splice.Splice.
func (x *Message) Decode(s *splice.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), MessageLen-magic.Len,
		MessageMagic); fails(e) {
		return
	}
	x.Return = &hidden.ReplyHeader{}
	s.ReadPubkey(&x.Address).
		ReadID(&x.ID).ReadID(&x.Re)
	hidden.ReadRoutingHeader(s, &x.Return.RoutingHeaderBytes).
		ReadCiphers(&x.Return.Ciphers).
		ReadNonces(&x.Return.Nonces).
		ReadBytes(&x.Payload)
	return
}

// Encode a Message into the next bytes of a splice.Splice.
func (x *Message) Encode(s *splice.Splice) (e error) {
	log.T.F("encoding %s %x %x %v %s", reflect.TypeOf(x),
		x.ID, x.Re, x.Address, spew.Sdump(x.Forward, x.Return,
			x.Payload.ToBytes()),
	)
	hidden.WriteRoutingHeader(s, x.Forward.RoutingHeaderBytes)
	start := s.GetCursor()
	s.Magic(MessageMagic).
		Pubkey(x.Address).
		ID(x.ID).ID(x.Re)
	hidden.WriteRoutingHeader(s, x.Return.RoutingHeaderBytes).
		Ciphers(x.Return.Ciphers).
		Nonces(x.Return.Nonces).
		Bytes(x.Payload)
	for i := range x.Forward.Ciphers {
		blk := ciph.BlockFromHash(x.Forward.Ciphers[i])
		log.D.F("encrypting %s", x.Forward.Ciphers[i])
		ciph.Encipher(blk, x.Forward.Nonces[i], s.GetFrom(start))
	}
	return
}

func Gen() codec.Codec { return &Message{} }

// Unwrap is a no-op because there is no onion inside a Message, only the reply parameters.
func (x *Message) Unwrap() interface{} { return nil }

// Handle is the relay logic for an engine handling a Message.
func (x *Message) Handle(s *splice.Splice, p ont.Onion, ng ont.Ngin) (e error) {
	// The hook in Pending should be the delivery of the reply to the client handler.
	_, e = ng.Pending().ProcessAndDelete(x.ID, x, s.GetAll())
	return
}

// Len returns the length of this Message.
func (x *Message) Len() int {

	codec.MustNotBeNil(x)

	return MessageLen + x.Payload.Len()
}

// Magic is the identifying 4 byte string indicating a Message follows.
func (x *Message) Magic() string { return MessageMagic }

// Wrap is a no-op because there is no further onion inside a Message, only reply parameters.
func (x *Message) Wrap(inner ont.Onion) {}

func init() { reg.Register(MessageMagic, Gen) }
