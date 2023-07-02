// Package adpeer provides a message type that provides the base information, identity key and relay rate for an Indra relay.
package adpeer

import (
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/crypto/sha256"
	"github.com/indra-labs/indra/pkg/engine/coding"
	"github.com/indra-labs/indra/pkg/engine/magic"
	"github.com/indra-labs/indra/pkg/onions/adintro"
	"github.com/indra-labs/indra/pkg/onions/adproto"
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
	Magic = "pead"
	Len   = adproto.Len +
		slice.Uint32Len
)

// Ad stores a specification for the fee rate and existence of a peer.
type Ad struct {
	adproto.Ad
	RelayRate uint32
}

var _ coding.Codec = &Ad{}

// New ...
func New(id nonce.ID, key *crypto.Prv, relayRate uint32,
	expiry time.Time) (sv *Ad) {

	s := splice.New(adintro.Len)
	k := crypto.DerivePub(key)
	Splice(s, id, k, relayRate, expiry)
	hash := sha256.Single(s.GetUntil(s.GetCursor()))
	var e error
	var sign crypto.SigBytes
	if sign, e = crypto.Sign(key, hash); fails(e) {
		return nil
	}
	sv = &Ad{
		Ad: adproto.Ad{
			ID:     id,
			Key:    k,
			Expiry: time.Now().Add(adproto.TTL),
			Sig:    sign,
		},
		RelayRate: relayRate,
	}
	if e = sv.Sign(key); fails(e) {
		return nil
	}
	return
}

func (x *Ad) Decode(s *splice.Splice) (e error) {
	s.ReadID(&x.ID).
		ReadPubkey(&x.Key).
		ReadUint32(&x.RelayRate).
		ReadTime(&x.Expiry).
		ReadSignature(&x.Sig)
	return
}

func (x *Ad) Encode(s *splice.Splice) (e error) {
	log.T.S("encoding", reflect.TypeOf(x), x)
	x.Splice(s)
	return
}

func (x *Ad) GetOnion() interface{} { return nil }

func (x *Ad) Len() int { return Len }

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

func (x *Ad) Splice(s *splice.Splice) {
	x.SpliceNoSig(s)
	s.Signature(x.Sig)
}

func (x *Ad) SpliceNoSig(s *splice.Splice) {
	Splice(s, x.ID, x.Key, x.RelayRate, x.Expiry)
}

func (x *Ad) Validate() (valid bool) {
	s := splice.New(adintro.Len - magic.Len)
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

func Splice(s *splice.Splice, id nonce.ID, key *crypto.Pub,
	relayRate uint32, expiry time.Time) {

	s.Magic(Magic).
		ID(id).
		Pubkey(key).
		Uint32(relayRate).
		Time(expiry)
}

func init() { reg.Register(Magic, peerAdGen) }

func peerAdGen() coding.Codec { return &Ad{} }
