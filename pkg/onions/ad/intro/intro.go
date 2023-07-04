// Package intro defines a message type that provides information about an introduction point for a hidden service.
package intro

import (
	"github.com/indra-labs/indra/pkg/onions/ad"
	"github.com/indra-labs/indra/pkg/onions/reg"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"net/netip"
	"reflect"
	"time"

	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/crypto/sha256"
	"github.com/indra-labs/indra/pkg/engine/coding"
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
	Len   = ad.Len + splice.AddrLen + 1 + slice.Uint16Len + slice.Uint32Len
)

// Ad is an Intro message that signals that a hidden service can be accessed from
// a given relay at a given address.
//
// todo: needs to be plural too!
type Ad struct {
	ad.Ad
	AddrPort  *netip.AddrPort // Introducer address.
	Port      uint16          // Well known port of protocol available.
	RelayRate uint32          // mSat/Mb
}

var _ coding.Codec = &Ad{}

// Decode an Ad out of the next bytes of a splice.Splice.
func (x *Ad) Decode(s *splice.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), Len-magic.Len,
		Magic); fails(e) {

		return
	}
	s.
		ReadID(&x.ID).
		ReadPubkey(&x.Key).
		ReadAddrPort(&x.AddrPort).
		ReadUint32(&x.RelayRate).
		ReadUint16(&x.Port).
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
//
// todo: plural.
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
	IntroSplice(s, x.ID, x.Key, x.AddrPort, x.RelayRate, x.Port, x.Expiry)
}

// Validate checks the signature matches the public key of the Ad.
func (x *Ad) Validate() bool {
	s := splice.New(Len - magic.Len)
	x.SpliceNoSig(s)
	hash := sha256.Single(s.GetUntil(s.GetCursor()))
	key, e := x.Sig.Recover(hash)
	if fails(e) {
		return false
	}
	if key.Equals(x.Key) && x.Expiry.After(time.Now()) {
		return true
	}
	return false
}

// IntroSplice creates the message part up to the signature for an Ad.
func IntroSplice(
	s *splice.Splice,
	id nonce.ID,
	key *crypto.Pub,
	ap *netip.AddrPort,
	relayRate uint32,
	port uint16,
	expires time.Time,
) {

	s.Magic(Magic).
		ID(id).
		Pubkey(key).
		AddrPort(ap).
		Uint32(relayRate).
		Uint16(port).
		Time(expires)
}

// New creates a new Ad and signs it.
func New(
	id nonce.ID,
	key *crypto.Prv,
	ap *netip.AddrPort,
	relayRate uint32,
	port uint16,
	expires time.Time,
) (in *Ad) {

	pk := crypto.DerivePub(key)

	in = &Ad{
		Ad: ad.Ad{
			ID:     id,
			Key:    pk,
			Expiry: expires,
		},
		AddrPort:  ap,
		RelayRate: relayRate,
		Port:      port,
	}
	s := splice.New(in.Len())
	in.SpliceNoSig(s)
	hash := sha256.Single(s.GetUntil(s.GetCursor()))
	var e error
	if in.Sig, e = crypto.Sign(key, hash); fails(e) {
		return nil
	}
	return
}

func init() { reg.Register(Magic, Gen) }

func Gen() coding.Codec { return &Ad{} }
