// Package intro defines a message type that provides information about an introduction point for a hidden service.
package intro

import (
	"github.com/indra-labs/indra/pkg/codec"
	"github.com/indra-labs/indra/pkg/codec/ad"
	"github.com/indra-labs/indra/pkg/codec/reg"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"time"

	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/engine/magic"
	"github.com/indra-labs/indra/pkg/util/slice"
	"github.com/indra-labs/indra/pkg/util/splice"
)

var (
	log   = log2.GetLogger()
	fails = log.E.Chk
)

const (
	Magic = "inad"
	Len   = ad.Len + crypto.PubKeyLen + slice.Uint16Len + slice.Uint32Len
)

// Ad is an Intro message that signals that a hidden service can be accessed from
// a given relay identifiable by its public key.
type Ad struct {

	// Embed ad.Ad for the common fields
	ad.Ad

	// Introducer is the key of the node that can forward a Route message to help
	// establish a connection to a hidden service.
	Introducer *crypto.Pub
	// Port is the well known port of protocol available.
	Port uint16

	// Rate for accessing the hidden service (covers the hidden service routing
	// header relaying).
	RelayRate uint32
}

var _ codec.Codec = &Ad{}

// Decode an Ad out of the next bytes of a splice.Splice.
func (x *Ad) Decode(s *splice.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), Len-magic.Len,
		Magic); fails(e) {

		return
	}
	s.
		ReadID(&x.ID).
		ReadPubkey(&x.Key).
		ReadPubkey(&x.Introducer).
		ReadUint32(&x.RelayRate).
		ReadUint16(&x.Port).
		ReadTime(&x.Expiry).
		ReadSignature(&x.Sig)
	return
}

// Encode an Ad into the next bytes of a splice.Splice.
func (x *Ad) Encode(s *splice.Splice) (e error) {
	x.Splice(s)
	return
}

// Unwrap returns nil because there is no onion inside.
func (x *Ad) Unwrap() interface{} { return nil }

// Len returns the length of the binary encoded Ad.
func (x *Ad) Len() int { return Len }

// Magic is the identifier indicating an Ad is encoded in the following bytes.
func (x *Ad) Magic() string { return Magic }

// Splice serializes an Ad into a splice.Splice.
func (x *Ad) Splice(s *splice.Splice) {
	x.SpliceNoSig(s)
	s.Signature(x.Sig)
}

// SpliceNoSig serializes the Ad but stops at the signature.
func (x *Ad) SpliceNoSig(s *splice.Splice) {
	Splice(s, x.ID, x.Key, x.Introducer, x.RelayRate, x.Port, x.Expiry)
}

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
func (x *Ad) Validate() bool {
	s := splice.New(x.Len())
	x.SpliceNoSig(s)
	return x.Sig.MatchesPubkey(s.GetUntilCursor(), x.Key) && x.Expiry.After(time.Now())
}

// Splice creates the message part up to the signature for an Ad.
func Splice(
	s *splice.Splice,
	id nonce.ID,
	key *crypto.Pub,
	introducer *crypto.Pub,
	relayRate uint32,
	port uint16,
	expires time.Time,
) {

	s.Magic(Magic).
		ID(id).
		Pubkey(key).
		Pubkey(introducer).
		Uint32(relayRate).
		Uint16(port).
		Time(expires)
}

// New creates a new Ad and signs it.
func New(
	id nonce.ID,
	key *crypto.Prv,
	introducer *crypto.Pub,
	relayRate uint32,
	port uint16,
	expires time.Time,
) (introAd *Ad) {

	pk := crypto.DerivePub(key)

	introAd = &Ad{
		Ad: ad.Ad{
			ID:     id,
			Key:    pk,
			Expiry: expires,
		},
		Introducer: introducer,
		RelayRate:  relayRate,
		Port:       port,
	}
	log.T.S("services ad", introAd)
	if e := introAd.Sign(key); fails(e) {
		return nil
	}
	log.T.S("signed", introAd)
	return

}

func init() { reg.Register(Magic, Gen) }

// Gen is a factory function for an Ad.
func Gen() codec.Codec { return &Ad{} }
