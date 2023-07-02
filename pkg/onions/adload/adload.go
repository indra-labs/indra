// Package adload provides a message type that provides information about the current load level of a node identified by its public key.
package adload

import (
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/crypto/sha256"
	"github.com/indra-labs/indra/pkg/engine/coding"
	"github.com/indra-labs/indra/pkg/engine/magic"
	"github.com/indra-labs/indra/pkg/onions/adproto"
	"github.com/indra-labs/indra/pkg/onions/reg"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/splice"
	"reflect"
	"time"
)

var (
	log   = log2.GetLogger()
	fails = log.E.Chk
)

const (
	Magic = "load"
	Len   = adproto.Len + 1
)

// Ad stores a specification for the fee rate and existence of a peer.
type Ad struct {
	adproto.Ad
	Load byte
}

var _ coding.Codec = &Ad{}

// New ...
func New(id nonce.ID, key *crypto.Prv, load byte,
	expiry time.Time) (sv *Ad) {

	k := crypto.DerivePub(key)
	sv = &Ad{
		Ad: adproto.Ad{
			ID:     id,
			Key:    k,
			Expiry: expiry,
		},
		Load: load,
	}
	s := splice.New(sv.Len())
	sv.SpliceNoSig(s)
	hash := sha256.Single(s.GetUntil(s.GetCursor()))
	var e error
	if sv.Sig, e = crypto.Sign(key, hash); fails(e) {
		return nil
	}
	return
}

func (x *Ad) Decode(s *splice.Splice) (e error) {
	s.ReadID(&x.ID).
		ReadPubkey(&x.Key).
		ReadByte(&x.Load).
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

func (x *Ad) Splice(s *splice.Splice) {
	x.SpliceNoSig(s)
	s.Signature(x.Sig)
}

func (x *Ad) SpliceNoSig(s *splice.Splice) {
	Splice(s, x.ID, x.Key, x.Load, x.Expiry)
}

func (x *Ad) Validate() (valid bool) {
	s := splice.New(x.Len() - magic.Len)
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

func Splice(s *splice.Splice, id nonce.ID, key *crypto.Pub,
	load byte, expiry time.Time) {

	s.Magic(Magic).
		ID(id).
		Pubkey(key).
		Byte(load).
		Time(expiry)
}

func init() { reg.Register(Magic, peerAdGen) }

func peerAdGen() coding.Codec { return &Ad{} }
