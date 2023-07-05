// Package services provides a message type for advertising what kinds of exit services a peer provides to clients, including the port number and the cost per megabyte of data.
package services

import (
	"fmt"
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
	"time"
)

var (
	log   = log2.GetLogger()
	fails = log.E.Chk
)

const (
	Magic      = "svad"
	ServiceLen = slice.Uint16Len + slice.Uint32Len
)

type Service struct {
	Port      uint16
	RelayRate uint32
}

// Ad stores a specification for the fee rate and the service port, which
// must be a well known port to match with a type of service, eg 80 for web, 53
// for DNS, etc. These are also attached to the PeerAd entry via concatenating
// "/service/N" where N is the index of the entry. A zero value at an index
// signals to stop scanning for more subsequent values.
type Ad struct {

	// Embed ad.Ad for the common fields
	ad.Ad

	// Services available on the relay identified by the public key.
	Services []Service
}

var _ codec.Codec = &Ad{}

// New creates a new Ad and signs it.
func New(id nonce.ID, key *crypto.Prv, services []Service,
	expiry time.Time) (sv *Ad) {

	k := crypto.DerivePub(key)
	sv = &Ad{
		Ad: ad.Ad{
			ID:     id,
			Key:    k,
			Expiry: expiry,
		},
		Services: services,
	}
	s := splice.New(intro.Len)
	sv.SpliceNoSig(s)
	hash := sha256.Single(s.GetUntil(s.GetCursor()))
	var e error
	if sv.Sig, e = crypto.Sign(key, hash); fails(e) {
		return nil
	}
	return
}

// Decode an Ad out of the next bytes of a splice.Splice.
func (x *Ad) Decode(s *splice.Splice) (e error) {
	var i, count uint32
	s.ReadID(&x.ID).
		ReadPubkey(&x.Key).
		ReadUint32(&count)
	x.Services = make([]Service, count)
	for ; i < count; i++ {
		s.ReadUint16(&x.Services[i].Port).
			ReadUint32(&x.Services[i].RelayRate)
	}
	s.
		ReadTime(&x.Expiry).
		ReadSignature(&x.Sig)
	return
}

// Encode an Ad into the next bytes of a splice.Splice.
func (x *Ad) Encode(s *splice.Splice) (e error) {
	x.Splice(s)
	return
}

// GetOnion returns nil because there is no onion inside.
func (x *Ad) GetOnion() interface{} { return nil }

// Len returns the length of the binary encoded Ad.
//
// This gives different values depending on how many services are listed.
func (x *Ad) Len() int { return ad.Len + len(x.Services)*ServiceLen + slice.Uint32Len }

// Magic is the identifier indicating an Ad is encoded in the following bytes.
func (x *Ad) Magic() string { return "" }

// Sign the Ad with the provided private key. It must match the embedded ad.Ad Key.
func (x *Ad) Sign(prv *crypto.Prv) (e error) {
	s := splice.New(x.Len())
	if e = x.Encode(s); fails(e) {
		return
	}
	var b []byte
	if b, e = prv.Sign(s.GetUntil(s.GetCursor())); fails(e) {
		return
	}
	if len(b) != crypto.SigLen {
		return fmt.Errorf("signature incorrect length, got %d expected %d",
			len(b), crypto.SigLen)
	}
	copy(x.Sig[:], b)
	return nil
}

// Splice serializes an Ad into a splice.Splice.
func (x *Ad) Splice(s *splice.Splice) {
	x.SpliceNoSig(s)
	s.Signature(x.Sig)
}

// SpliceNoSig serializes the Ad but stops at the signature.
func (x *Ad) SpliceNoSig(s *splice.Splice) {
	ServiceSplice(s, x.ID, x.Key, x.Services, x.Expiry)
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

// ServiceSplice creates the message part up to the signature for an Ad.
func ServiceSplice(s *splice.Splice, id nonce.ID, key *crypto.Pub, services []Service, expiry time.Time) {

	s.Magic(Magic).
		ID(id).
		Pubkey(key).
		Uint32(uint32(len(services)))
	for i := range services {
		s.
			Uint16(services[i].Port).
			Uint32(services[i].RelayRate)
	}
	s.Time(expiry)
}

func init() { reg.Register(Magic, Gen) }

// Gen is the factory function for an Ad.
func Gen() codec.Codec { return &Ad{} }
