// Package addresses defines the message format that provides the network multi-address of a peer with a given public identity key.
package addresses

import (
	"fmt"
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/crypto/sha256"
	"github.com/indra-labs/indra/pkg/engine/coding"
	"github.com/indra-labs/indra/pkg/engine/magic"
	"github.com/indra-labs/indra/pkg/onions/ad"
	"github.com/indra-labs/indra/pkg/onions/ad/intro"
	"github.com/indra-labs/indra/pkg/onions/reg"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/slice"
	"github.com/indra-labs/indra/pkg/util/splice"
	"net/netip"
	"time"
)

var (
	log   = log2.GetLogger()
	fails = log.E.Chk
)

const (
	Magic   = "svad"
	AddrLen = splice.AddrLen
)

// Ad stores a specification for the fee rate and the service port, which
// must be a well known port to match with a type of service, eg 80 for web, 53
// for DNS, etc. These are also attached to the PeerAd entry via concatenating
// "/service/N" where N is the index of the entry. A zero value at an index
// signals to stop scanning for more subsequent values.
type Ad struct {
	ad.Ad
	Addresses []*netip.AddrPort
}

var _ coding.Codec = &Ad{}

// New creates a new addresses.Ad.
func New(id nonce.ID, key *crypto.Prv, addrs []*netip.AddrPort,
	expiry time.Time) (sv *Ad) {

	k := crypto.DerivePub(key)
	sv = &Ad{
		Ad: ad.Ad{
			ID:     id,
			Key:    k,
			Expiry: expiry,
		},
		Addresses: addrs,
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

// Decode a splice.Splice's next bytes into an Ad.
func (x *Ad) Decode(s *splice.Splice) (e error) {
	var i, count uint32
	s.ReadID(&x.ID).
		ReadPubkey(&x.Key).
		ReadUint32(&count)
	x.Addresses = make([]*netip.AddrPort, count)
	for ; i < count; i++ {
		addy := &netip.AddrPort{}
		s.ReadAddrPort(&addy)
		x.Addresses[i] = addy
	}
	s.
		ReadTime(&x.Expiry).
		ReadSignature(&x.Sig)
	return
}

// Encode an Ad into a splice.Splice's next bytes.
func (x *Ad) Encode(s *splice.Splice) (e error) {
	x.Splice(s)
	return
}

// GetOnion returns nothing because there isn't an onion inside an Ad.
func (x *Ad) GetOnion() interface{} { return nil }

// Len returns the length of bytes required to encode the Ad, based on the number
// of Addresses inside it.
func (x *Ad) Len() int { return ad.Len + len(x.Addresses)*AddrLen + slice.Uint32Len + 2 }

// Magic bytes that identify this message
func (x *Ad) Magic() string { return Magic }

// Sign the encoded form of the bytes in order to authorise it.
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

// Splice together an Ad.
func (x *Ad) Splice(s *splice.Splice) {
	x.SpliceNoSig(s)
	s.Signature(x.Sig)
}

// SpliceNoSig splices until the signature.
func (x *Ad) SpliceNoSig(s *splice.Splice) {
	Splice(s, x.ID, x.Key, x.Addresses, x.Expiry)
}

// Validate checks that the signature matches the public key.
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

// Splice is a function that serializes the parts of an Ad.
func Splice(s *splice.Splice, id nonce.ID, key *crypto.Pub,
	addrs []*netip.AddrPort, expiry time.Time) {

	s.Magic(Magic).
		ID(id).
		Pubkey(key).
		Uint32(uint32(len(addrs)))
	for i := range addrs {
		s.AddrPort(addrs[i])
	}
	s.Time(expiry)
}

func init() { reg.Register(Magic, Gen) }

func Gen() coding.Codec { return &Ad{} }
