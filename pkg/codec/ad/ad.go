// Package ad is an abstract message type that composes the common elements of all ads - nonce ID, public key (identity), expiry and signature.
//
// The concrete ad types are in subfolders of this package.
package ad

import (
	"github.com/indra-labs/indra/pkg/codec"
	"github.com/indra-labs/indra/pkg/codec/reg"
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/crypto/sha256"
	"github.com/indra-labs/indra/pkg/engine/magic"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/slice"
	"github.com/indra-labs/indra/pkg/util/splice"
	"reflect"
	"time"
)

var (
	log   = log2.GetLogger()
	fails = log.E.Chk
)

const (
	Magic = "prad"
	Len   = magic.Len +
		nonce.IDLen +
		crypto.PubKeyLen +
		slice.Uint64Len +
		crypto.SigLen
	TTL = time.Hour
)

// Ad is an abstract message that isn't actually used, but rather is composed
// into the rest being the common fields to all ads.
type Ad struct {

	// ID is a random number generated for each new add that practically guarantees
	// no repeated hash for a message to be signed with the same peer identity key.
	ID nonce.ID

	// Key is the public identity key of the peer.
	Key *crypto.Pub

	// Expiry is the time after which this ad is considered no longer current.
	Expiry time.Time

	// Sig is the signature, which must match the Key above.
	Sig crypto.SigBytes
}

// Decode unpacks a binary encoded form of the Ad and populates itself.
func (x *Ad) Decode(s *splice.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), Len-magic.Len,
		Magic); fails(e) {

		return
	}
	s.ReadID(&x.ID).
		ReadPubkey(&x.Key).
		ReadTime(&x.Expiry).
		ReadSignature(&x.Sig)
	return
}

// Encode the Ad into a splice.Splice for wire or storage.
func (x *Ad) Encode(s *splice.Splice) (e error) {
	log.T.S("encoding", reflect.TypeOf(x), x)
	x.Splice(s)
	return
}

// GetOnion returns nil because there is no onion inside an Ad.
func (x *Ad) GetOnion() interface{} { return nil }

// Len is the number of bytes required for the binary encoded form of an Ad.
func (x *Ad) Len() int { return Len }

// Magic is the identifying 4 byte string used to mark the beginning of a message
// and designate the type.
func (x *Ad) Magic() string { return Magic }

// Splice the Ad into a splice.Splice.
func (x *Ad) Splice(s *splice.Splice) {
	x.SpliceNoSig(s)
	s.Signature(x.Sig)
}

// SpliceNoSig encodes the message but stops at the signature.
func (x *Ad) SpliceNoSig(s *splice.Splice) {
	s.Magic(Magic).
		ID(x.ID).
		Pubkey(x.Key).
		Time(x.Expiry)
}

// Validate returns true if the signature matches the public key.
func (x *Ad) Validate() bool {
	s := splice.New(Len)
	x.SpliceNoSig(s)
	hash := sha256.Single(s.GetUntil(s.GetCursor()))
	key, e := x.Sig.Recover(hash)
	if fails(e) {
		return false
	}
	if key.Equals(x.Key) {
		return true
	}
	return false
}

// New creates a new Ad and signs it with the provided private key.
func New(id nonce.ID, key *crypto.Prv,
	expiry time.Time) (protoAd *Ad) {

	pub := crypto.DerivePub(key)
	protoAd = &Ad{
		ID:     id,
		Key:    pub,
		Expiry: expiry,
	}
	s := splice.New(Len - magic.Len)
	protoAd.SpliceNoSig(s)
	hash := sha256.Single(s.GetUntil(s.GetCursor()))
	var e error
	if protoAd.Sig, e = crypto.Sign(key, hash); fails(e) {
		return nil
	}
	return
}

func Gen() codec.Codec { return &Ad{} }

func init() { reg.Register(Magic, Gen) }
