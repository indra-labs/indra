// Package session provides an onion message type that delivers the two session private keys to be associated with a session, for which the hash of the secrets was used as the payment preimage for starting a session.
//
// Topping up sessions does not require following up with this message as the handler finds the session and adjusts the balance according to the payment.
package session

import (
	"github.com/indra-labs/indra/pkg/codec"
	"github.com/indra-labs/indra/pkg/codec/onion/cores/end"
	"github.com/indra-labs/indra/pkg/codec/ont"
	"github.com/indra-labs/indra/pkg/codec/reg"
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/crypto/sha256"
	"github.com/indra-labs/indra/pkg/engine/magic"
	"github.com/indra-labs/indra/pkg/engine/sess"
	"github.com/indra-labs/indra/pkg/engine/sessions"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/splice"
	"reflect"
)

var (
	log   = log2.GetLogger()
	fails = log.E.Chk
)

const (
	Magic = "sess"
	Len   = magic.Len + nonce.IDLen + crypto.PrvKeyLen*2
)

// Session is the delivery of the two client defined private keys that a relay
// uses to identify and encrypt/decrypt messages related to a session.
type Session struct {

	// ID is an identifier only used by the client.
	ID nonce.ID

	// Hop is the position in the circuit this session is used, also private to the client.
	Hop byte

	// Header and Payload are the crypto.Keys based on the two private keys used for
	// the session (and hashed into the preimage).
	Header, Payload *crypto.Keys

	// Onion is the rest of the message, which usually will be some number more
	// Session/Crypt layers.
	ont.Onion
}

// New creates a new Session, including generating new keys for it.
func New(hop byte) (x ont.Onion) {
	var e error
	var hdr, pld *crypto.Keys
	if hdr, pld, e = crypto.Generate2Keys(); fails(e) {
		return
	}
	return &Session{
		ID:      nonce.NewID(),
		Hop:     hop,
		Header:  hdr,
		Payload: pld,
		Onion:   end.NewEnd(),
	}
}

// Account for the Session message - in this case, this is the one exception to
// the rule all messages must have sessions to relay, which is usually what is
// inside Session Onion fields as well up to the return confirmation layer.
func (x *Session) Account(res *sess.Data, sm *sess.Manager, s *sessions.Data,
	last bool) (skip bool, sd *sessions.Data) {
	return
}

// Decode a Session from a provided splice.Splice.
func (x *Session) Decode(s *splice.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), Len-magic.Len,
		Magic); fails(e) {
		return
	}
	var h, p crypto.Prv
	hdr, pld := &h, &p
	s.
		ReadID(&x.ID).
		ReadPrvkey(&hdr).
		ReadPrvkey(&pld)
	x.Header, x.Payload = crypto.MakeKeys(hdr), crypto.MakeKeys(pld)
	return
}

// Encode a Session into the next bytes of a splice.Splice.
func (x *Session) Encode(s *splice.Splice) (e error) {
	log.T.S("encoding", reflect.TypeOf(x),
		x.ID, x.Hop, x.Header, x.Payload,
	)
	return x.Onion.Encode(s.Magic(Magic).
		ID(x.ID).
		Prvkey(x.Header.Prv).
		Prvkey(x.Payload.Prv),
	)
}

// Unwrap returns the onion inside this Session.
func (x *Session) Unwrap() interface{} { return x.Onion }

// Handle is the relay logic for an engine handling a Session message.
//
// If the relay finds a pending payment it will forward the next part of the
// onion, otherwise not. This is the only onion that relaying isn't charged for
// because it establishes the account.
func (x *Session) Handle(s *splice.Splice, p ont.Onion, ng ont.Ngin) (e error) {
	log.T.F("incoming session %s", x.PreimageHash())
	pi := ng.Mgr().FindPendingPreimage(x.PreimageHash())
	if pi != nil {
		// We need to delete this first in case somehow two such messages arrive
		// at the same time, and we end up with duplicate
		ng.Mgr().DeletePendingPayment(pi.Preimage)
		log.D.F("adding session %s to %s", pi.ID,
			ng.Mgr().GetLocalNodeAddressString())
		ng.Mgr().AddSession(sessions.NewSessionData(pi.ID,
			ng.Mgr().GetLocalNode(), pi.Amount, x.Header, x.Payload, x.Hop))
		ng.HandleMessage(splice.BudgeUp(s), nil)
	} else {
		log.E.Ln("dropping session message without payment")
	}
	return
}

// Len returns the length of this Session message.
func (x *Session) Len() int {

	codec.MustNotBeNil(x)

	return Len + x.Onion.Len()
}

// Magic is the identifying 4 byte string indicating a Session message follows.
func (x *Session) Magic() string { return Magic }

// PreimageHash returns the preimage to use in a LN payment to associate to this
// session.
func (x *Session) PreimageHash() sha256.Hash {
	h, p := x.Header.Prv.ToBytes(), x.Payload.Prv.ToBytes()
	return sha256.Single(append(h[:], p[:]...))
}

// Wrap puts another onion inside this Reverse onion.
func (x *Session) Wrap(inner ont.Onion) { x.Onion = inner }

func init() { reg.Register(Magic, Gen) }

// Gen is a factory function for a Session.
func Gen() codec.Codec { return &Session{} }
