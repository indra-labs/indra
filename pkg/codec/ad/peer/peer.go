// Package peer provides a message type that provides the base information, identity key and relay rate for an Indra relay.
package peer

import (
	"github.com/indra-labs/indra/pkg/codec"
	"github.com/indra-labs/indra/pkg/codec/ad"
	"github.com/indra-labs/indra/pkg/codec/ad/intro"
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
	Magic = "pead"
	Len   = ad.Len +
		slice.Uint32Len
)

// Ad stores a specification for the relaying fee rate and existence of a peer.
type Ad struct {

	// Embed ad.Ad for the common fields
	ad.Ad

	// RelayRate is the fee for forwarding packets, mSAT/Mb (1024^3 bytes).
	RelayRate uint32
}

var _ codec.Codec = &Ad{}

// New creates a new Ad and signs it with the provided private key.
func New(id nonce.ID, key *crypto.Prv, relayRate uint32,
	expiry time.Time) (sv *Ad) {

	s := splice.New(intro.Len)
	k := crypto.DerivePub(key)
	Splice(s, id, k, relayRate, expiry)
	hash := sha256.Single(s.GetUntil(s.GetCursor()))
	var e error
	var sign crypto.SigBytes
	if sign, e = crypto.Sign(key, hash); fails(e) {
		return nil
	}
	sv = &Ad{
		Ad: ad.Ad{
			ID:     id,
			Key:    k,
			Expiry: time.Now().Add(ad.TTL),
			Sig:    sign,
		},
		RelayRate: relayRate,
	}
	if e = sv.Sign(key); fails(e) {
		return nil
	}
	return
}

// Decode an Ad out of the next bytes of a splice.Splice.
func (x *Ad) Decode(s *splice.Splice) (e error) {
	s.ReadID(&x.ID).
		ReadPubkey(&x.Key).
		ReadUint32(&x.RelayRate).
		ReadTime(&x.Expiry).
		ReadSignature(&x.Sig)
	return
}

// Encode an Ad into the next bytes of a splice.Splice.
func (x *Ad) Encode(s *splice.Splice) (e error) {
	log.T.S("encoding", reflect.TypeOf(x), x)
	x.Splice(s)
	return
}

// GetOnion returns nil because there is no onion inside.
func (x *Ad) GetOnion() interface{} { return nil }

// Len returns the length of the binary encoded Ad.
func (x *Ad) Len() int { return Len }

// Magic is the identifier indicating an Ad is encoded in the following bytes.
func (x *Ad) Magic() string { return "" }

func (x *Ad) Sign(prv *crypto.Prv) (e error) {
	s := splice.New(x.Len())
	x.SpliceNoSig(s)
	var b crypto.SigBytes
	if b, e = crypto.Sign(prv, sha256.Single(s.GetUntil(s.GetCursor()))); fails(e) {
		return
	}
	copy(x.Sig[:], b[:])
	return nil
}

// Splice serializes an Ad into a splice.Splice.
func (x *Ad) Splice(s *splice.Splice) {
	x.SpliceNoSig(s)
	s.Signature(x.Sig)
}

// SpliceNoSig serializes the Ad but stops at the signature.
func (x *Ad) SpliceNoSig(s *splice.Splice) {
	Splice(s, x.ID, x.Key, x.RelayRate, x.Expiry)
}

// Validate checks the signature matches the public key of the Ad.
func (x *Ad) Validate() (valid bool) {
	s := splice.New(intro.Len - magic.Len)
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

// Splice serializes an Ad into a splice.Splice.
func Splice(s *splice.Splice, id nonce.ID, key *crypto.Pub,
	relayRate uint32, expiry time.Time) {

	s.Magic(Magic).
		ID(id).
		Pubkey(key).
		Uint32(relayRate).
		Time(expiry)
}

func init() { reg.Register(Magic, Gen) }

func Gen() codec.Codec { return &Ad{} }
