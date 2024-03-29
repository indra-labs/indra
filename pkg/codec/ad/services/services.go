// Package services provides a message type for advertising what kinds of exit services a peer provides to clients, including the port number and the cost per megabyte of data.
package services

import (
	"git.indra-labs.org/dev/ind/pkg/codec"
	"git.indra-labs.org/dev/ind/pkg/codec/ad"
	"git.indra-labs.org/dev/ind/pkg/codec/reg"
	"git.indra-labs.org/dev/ind/pkg/crypto"
	"git.indra-labs.org/dev/ind/pkg/crypto/nonce"
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
	Magic = "svad"
	Len   = ad.Len +
		slice.Uint16Len +
		slice.Uint32Len
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
	expiry time.Time) (svcAd *Ad) {

	k := crypto.DerivePub(key)
	svcAd = &Ad{
		Ad: ad.Ad{
			ID:     id,
			Key:    k,
			Expiry: expiry,
		},
		Services: services,
	}
	log.T.S("services ad", svcAd)
	if e := svcAd.Sign(key); fails(e) {
		return nil
	}
	log.T.S("signed", svcAd)
	return
}

func (x *Ad) PubKey() (key *crypto.Pub) { return x.Key }
func (x *Ad) Fingerprint() (pf string)  { return x.Key.Fingerprint() }
func (x *Ad) Expired() (is bool)        { return x.Expiry.Before(time.Now()) }

func (x *Ad) GetID() (id peer.ID, e error) {
	return peer.IDFromPublicKey(x.Key)
}

// Decode an Ad out of the next bytes of a splice.Splice.
func (x *Ad) Decode(s *splice.Splice) (e error) {
	var i, count uint16
	s.ReadID(&x.ID).
		ReadPubkey(&x.Key).
		ReadUint16(&count)
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
//
// This gives different values depending on how many services are listed.
func (x *Ad) Len() int {

	codec.MustNotBeNil(x)

	return ad.Len + len(x.Services)*Len + slice.Uint32Len
}

// Magic is the identifier indicating an Ad is encoded in the following bytes.
func (x *Ad) Magic() string { return "" }

// Sign the Ad with the provided private key. It must match the embedded ad.Ad Key.
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
	return x.Sig.MatchesPubkey(s.GetUntilCursor(), x.Key) && x.Expiry.After(time.Now())
}

// SpliceNoSig serializes the Ad but stops at the signature.
func (x *Ad) SpliceNoSig(s *splice.Splice) {
	Splice(s, x.ID, x.Key, x.Services, x.Expiry)
}

// Splice creates the message part up to the signature for an Ad.
func Splice(s *splice.Splice, id nonce.ID, key *crypto.Pub, services []Service, expiry time.Time) {

	s.Magic(Magic).
		ID(id).
		Pubkey(key).
		Uint16(uint16(len(services)))
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
