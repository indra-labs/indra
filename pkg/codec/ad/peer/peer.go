// Package peer provides a message type that provides the base information, identity key and relay rate for an Indra relay.
package peer

import (
	"git.indra-labs.org/dev/ind/pkg/codec"
	"git.indra-labs.org/dev/ind/pkg/codec/ad"
	"git.indra-labs.org/dev/ind/pkg/codec/ad/intro"
	"git.indra-labs.org/dev/ind/pkg/codec/reg"
	"git.indra-labs.org/dev/ind/pkg/crypto"
	"git.indra-labs.org/dev/ind/pkg/crypto/nonce"
	"git.indra-labs.org/dev/ind/pkg/crypto/sha256"
	log2 "git.indra-labs.org/dev/ind/pkg/proc/log"
	"git.indra-labs.org/dev/ind/pkg/util/slice"
	"git.indra-labs.org/dev/ind/pkg/util/splice"
	"github.com/libp2p/go-libp2p/core/peer"
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

func (x *Ad) PubKey() (key *crypto.Pub) { return x.Key }
func (x *Ad) Fingerprint() (pf string)  { return x.Key.Fingerprint() }
func (x *Ad) Expired() (is bool)        { return x.Expiry.Before(time.Now()) }

func (x *Ad) GetID() (id peer.ID, e error) {
	return peer.IDFromPublicKey(x.Key)
}

// New creates a new Ad and signs it with the provided private key.
func New(id nonce.ID, key *crypto.Prv, relayRate uint32,
	expiry time.Time) (peerAd *Ad) {

	s := splice.New(intro.Len)
	k := crypto.DerivePub(key)
	Splice(s, id, k, relayRate, expiry)
	hash := sha256.Single(s.GetUntil(s.GetCursor()))
	var e error
	var sign crypto.SigBytes
	if sign, e = crypto.Sign(key, hash); fails(e) {
		return nil
	}
	peerAd = &Ad{
		Ad: ad.Ad{
			ID:     id,
			Key:    k,
			Expiry: time.Now().Add(time.Hour * 3),
			Sig:    sign,
		},
		RelayRate: relayRate,
	}
	log.T.S("peer ad", peerAd)
	if e = peerAd.Sign(key); fails(e) {
		return nil
	}
	log.T.S("signed", peerAd)
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

// Encode an Ad into a splice.Splice's next bytes. It is assumed the
// signature has been generated, or it would be an invalid Ad.
func (x *Ad) Encode(s *splice.Splice) (e error) {
	x.SpliceNoSig(s)
	s.Signature(x.Sig)
	return
}

// Unwrap returns nil because there is no onion inside.
func (x *Ad) Unwrap() interface{} { return nil }

// Len returns the length of the binary encoded Ad.
func (x *Ad) Len() int {

	codec.MustNotBeNil(x)

	return Len
}

// Magic is the identifier indicating an Ad is encoded in the following bytes.
func (x *Ad) Magic() string { return "" }

func (x *Ad) Sign(prv *crypto.Prv) (e error) {
	s := splice.New(x.Len())
	x.SpliceNoSig(s)
	var b []byte
	if b, e = prv.Sign(s.GetUntilCursor()); fails(e) {
		return
	}
	copy(x.Sig[:], b[:])
	return
}

// Validate checks the signature matches the public key of the Ad.
func (x *Ad) Validate() (valid bool) {
	s := splice.New(x.Len())
	x.SpliceNoSig(s)
	return x.Sig.MatchesPubkey(s.GetUntilCursor(), x.Key) &&
		x.Expiry.After(time.Now())
}

// SpliceNoSig serializes the Ad but stops at the signature.
func (x *Ad) SpliceNoSig(s *splice.Splice) {
	Splice(s, x.ID, x.Key, x.RelayRate, x.Expiry)
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
