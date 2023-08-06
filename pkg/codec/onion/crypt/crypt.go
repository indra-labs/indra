// Package crypt is an onion message layer which specifies that subsequent content will be encrypted.
//
// The cloaked receiver key, and the ephemeral per-message/per-packet "from" keys are intended to be single use only (generated via scalar multiplication with pairs of secrets).
//
// todo: note reference of this algorithm.
package crypt

import (
	"crypto/cipher"
	"github.com/indra-labs/indra/pkg/codec"
	"github.com/indra-labs/indra/pkg/codec/onion/cores/end"
	"github.com/indra-labs/indra/pkg/codec/onion/session"
	"github.com/indra-labs/indra/pkg/codec/ont"
	"github.com/indra-labs/indra/pkg/codec/reg"
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/ciph"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/engine/consts"
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
	CryptMagic = "cryp"
)

// Crypt is an encrypted message, and forms the "skins" of the onions.
type Crypt struct {

	// Depth is used with RoutingHeaders to indicate which of the 3 layers in a
	// ReverseCrypt section.
	Depth int

	// ToHeaderPub, ToPayloadPub are the public keys of the session.
	ToHeaderPub, ToPayloadPub *crypto.Pub

	// From is usually a one-time generated private key for which the public
	// counterpart combined with the recipient's private key generates the same
	// secret via ECDH.
	From *crypto.Prv

	// IV is the Initialization Vector for the AES-CTR encryption used in a Crypt.
	IV nonce.IV

	// The remainder here are for Decode.

	// Cloak is the obfuscated receiver key.
	Cloak crypto.CloakedPubKey

	// ToPriv is the private key the receiver knows.
	ToPriv *crypto.Prv

	// FromPub is the public key encoded into the Crypt header.
	FromPub *crypto.Pub

	// Onion contains the rest of the message.
	ont.Onion
}

// Account attaches the session, which is tied to the keys used in the crypt, to the pending result.
func (x *Crypt) Account(res *sess.Data, sm *sess.Manager, s *sessions.Data,
	last bool) (skip bool, sd *sessions.Data) {

	sd = sm.FindSessionByHeaderPub(x.ToHeaderPub)
	if sd == nil {
		return
	}
	res.Sessions = append(res.Sessions, sd)
	// The last hop needs no accounting as it's us!
	if last {
		res.Ret = sd.Header.Bytes
		res.Billable = append(res.Billable, sd.Header.Bytes)
	}
	return
}

// Decode a splice.Splice's next bytes into a Crypt.
func (x *Crypt) Decode(s *splice.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), consts.CryptLen-magic.Len,
		CryptMagic); fails(e) {

		return
	}
	s.ReadIV(&x.IV).ReadCloak(&x.Cloak).ReadPubkey(&x.FromPub)
	return
}

// Decrypt requires the prv.Pub to be located from the Cloak, using the FromPub
// key to derive the shared secret, and then decrypts the rest of the message.
func (x *Crypt) Decrypt(prk *crypto.Prv, s *splice.Splice) {
	ciph.Encipher(ciph.GetBlock(prk, x.FromPub, "decrypt crypt header"),
		x.IV, s.GetRest())
}

// Encode a Crypt into a splice.Splice's next bytes.
//
// The crypt renders the inner contents first and once complete returns and
// encrypts everything after the Crypt header.
func (x *Crypt) Encode(s *splice.Splice) (e error) {
	log.T.F("encoding %s %s %x %x", reflect.TypeOf(x),
		x.ToHeaderPub, x.From.ToBytes(), x.IV,
	)
	if x.ToHeaderPub == nil || x.From == nil {
		s.Advance(consts.CryptLen, "crypt")
		return
	}
	s.Magic(CryptMagic).
		IV(x.IV).Cloak(x.ToHeaderPub).Pubkey(crypto.DerivePub(x.From))
	// Then we can encrypt the message segment
	var blk cipher.Block
	if blk = ciph.GetBlock(x.From, x.ToHeaderPub,
		"crypt header"); fails(e) {

		panic(e)
	}
	start := s.GetCursor()
	end := s.Len()
	switch {
	case x.Depth == 0:
	case x.Depth > 0:
		end = start + (x.Depth-1)*consts.ReverseCryptLen
	default:
		panic("incorrect value for crypt sequence")
	}
	if x.Onion != nil {
		if e = x.Onion.Encode(s); fails(e) {
			return
		}
	}
	ciph.Encipher(blk, x.IV, s.GetRange(start, end))
	if end != s.Len() {
		if blk = ciph.GetBlock(x.From, x.ToPayloadPub,
			"crypt payload"); fails(e) {
			return
		}
		ciph.Encipher(blk, x.IV, s.GetFrom(end))
	}
	return e
}

// Unwrap returns the layers inside the crypt..
func (x *Crypt) Unwrap() interface{} { return x }

// Handle provides relay and accounting processing logic for receiving a Crypt
// message.
func (x *Crypt) Handle(s *splice.Splice, p ont.Onion, ng ont.Ngin) (e error) {

	// todo: get the payload key also, and read ahead for an offset, and apply

	hdr, _, _, identity := ng.Mgr().FindCloaked(x.Cloak)
	if hdr == nil {
		log.T.Ln("no matching key found from cloaked key")
		return
	}
	x.ToPriv = hdr
	x.Decrypt(hdr, s)
	if identity {
		if string(s.GetRest()[:magic.Len]) != session.Magic {
			log.T.Ln("dropping message due to identity key with" +
				" no following session")
			return e
		}
	}

	ng.HandleMessage(splice.BudgeUp(s), x)
	return e
}

// Len returns the length of bytes required to encode the Crypt.
func (x *Crypt) Len() int {

	codec.MustNotBeNil(x)

	return consts.CryptLen + x.Onion.Len()
}

// Magic bytes that identify this message
func (x *Crypt) Magic() string { return CryptMagic }

// Wrap inserts an onion inside a Crypt.
func (x *Crypt) Wrap(inner ont.Onion) { x.Onion = inner }

// New creates a new crypt message with an empty slot for more messages.
func New(toHdr, toPld *crypto.Pub, from *crypto.Prv, iv nonce.IV,
	depth int) ont.Onion {
	return &Crypt{
		Depth:        depth,
		ToHeaderPub:  toHdr,
		ToPayloadPub: toPld,
		From:         from,
		IV:           iv,
		Onion:        end.NewEnd(),
	}
}

// Gen is a factory function to generate an Crypt.
func Gen() codec.Codec { return &Crypt{} }

func init() { reg.Register(CryptMagic, Gen) }
