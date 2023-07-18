// Package load provides a message type that provides information about the current load level of a node identified by its public key.
package load

import (
	"github.com/indra-labs/indra/pkg/codec"
	"github.com/indra-labs/indra/pkg/codec/ad"
	"github.com/indra-labs/indra/pkg/codec/reg"
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/splice"
	"time"
)

var (
	log   = log2.GetLogger()
	fails = log.E.Chk
)

const (
	Magic = "load"
	Len   = ad.Len + 1
)

// Ad stores a specification for the fee rate and existence of a peer.
type Ad struct {

	// Embed ad.Ad for the common fields
	ad.Ad

	// Load is a value that represents utilisation as a value from 0 to 255.
	Load byte
}

var _ codec.Codec = &Ad{}

// New creates a new Ad.
func New(id nonce.ID, key *crypto.Prv, load byte,
	expiry time.Time) (loAd *Ad) {

	k := crypto.DerivePub(key)
	loAd = &Ad{
		Ad: ad.Ad{
			ID:     id,
			Key:    k,
			Expiry: expiry,
		},
		Load: load,
	}
	log.T.S("load ad", loAd)
	if e := loAd.Sign(key); fails(e) {
		return
	}
	log.T.S("signed", loAd)
	return
}

// Decode unpacks a binary encoded form of the Ad and populates itself.
func (x *Ad) Decode(s *splice.Splice) (e error) {
	s.ReadID(&x.ID).
		ReadPubkey(&x.Key).
		ReadByte(&x.Load).
		ReadTime(&x.Expiry).
		ReadSignature(&x.Sig)
	return
}

// Encode the Ad into a splice.Splice for wire or storage.
func (x *Ad) Encode(s *splice.Splice) (e error) {
	x.Splice(s)
	return
}

// Unwrap returns nil because there is no onion inside an Ad.
func (x *Ad) Unwrap() interface{} { return nil }

// Len is the number of bytes required for the binary encoded form of an Ad.
func (x *Ad) Len() int { return Len }

// Magic is the identifying 4 byte string used to mark the beginning of a message
// and designate the type.
func (x *Ad) Magic() string { return "" }

// Splice the Ad into a splice.Splice.
func (x *Ad) Splice(s *splice.Splice) {
	x.SpliceNoSig(s)
	s.Signature(x.Sig)
}

// SpliceNoSig encodes the message but stops at the signature.
func (x *Ad) SpliceNoSig(s *splice.Splice) {
	Splice(s, x.ID, x.Key, x.Load, x.Expiry)
}

func (x *Ad) Sign(prv *crypto.Prv) (e error) {
	s := splice.New(x.Len())
	x.SpliceNoSig(s)
	log.T.S("message", s.GetUntilCursor().ToBytes())
	var b []byte
	if b, e = prv.Sign(s.GetUntil(s.GetCursor())); fails(e) {
		return
	}
	log.T.S("signature", b)
	copy(x.Sig[:], b)
	return nil
}

// Validate returns true if the signature matches the public key.
func (x *Ad) Validate() (valid bool) {
	s := splice.New(x.Len())
	x.SpliceNoSig(s)
	return x.Sig.MatchesPubkey(s.GetUntilCursor(), x.Key) &&
		x.Expiry.After(time.Now())
}

// Splice the Ad into a splice.Splice.
func Splice(s *splice.Splice, id nonce.ID, key *crypto.Pub,
	load byte, expiry time.Time) {

	s.Magic(Magic).
		ID(id).
		Pubkey(key).
		Byte(load).
		Time(expiry)
}

func init() { reg.Register(Magic, Gen) }

// Gen is a factory function for an Ad.
func Gen() codec.Codec { return &Ad{} }
