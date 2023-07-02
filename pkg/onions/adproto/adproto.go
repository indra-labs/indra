// Package adproto is an abstract message type that composes the common elements of all ads - nonce ID, public key (identity), expiry and signature.
package adproto

import (
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/crypto/sha256"
	"github.com/indra-labs/indra/pkg/engine/coding"
	"github.com/indra-labs/indra/pkg/engine/magic"
	"github.com/indra-labs/indra/pkg/onions/reg"
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

// Ad entries are stored with an index generated by concatenating the bytes of
// the public key with a string path "/address/N" where N is the index of the
// address. This means hidden service introducers for values over zero. Hidden
// services have no value in the zero index, which is "<hash>/address/0".
type Ad struct {
	ID     nonce.ID // To ensure no repeating message
	Key    *crypto.Pub
	Expiry time.Time
	Sig    crypto.SigBytes
}

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

func (x *Ad) Encode(s *splice.Splice) (e error) {
	log.T.S("encoding", reflect.TypeOf(x), x)
	x.Splice(s)
	return
}

func (x *Ad) GetOnion() interface{} { return x }

func (x *Ad) Len() int { return Len }

func (x *Ad) Magic() string { return Magic }

func (x *Ad) Splice(s *splice.Splice) {
	x.SpliceNoSig(s)
	s.Signature(x.Sig)
}

func (x *Ad) SpliceNoSig(s *splice.Splice) {
	s.Magic(Magic).
		ID(x.ID).
		Pubkey(x.Key).
		Time(x.Expiry)
}

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

func protGen() coding.Codec { return &Ad{} }

func init() { reg.Register(Magic, protGen) }
