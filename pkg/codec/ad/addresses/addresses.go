// Package addresses defines the message format that provides the network multi-address of a peer with a given public identity key.
package addresses

import (
	"github.com/indra-labs/indra/pkg/codec"
	"github.com/indra-labs/indra/pkg/codec/ad"
	"github.com/indra-labs/indra/pkg/codec/reg"
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
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
	Magic = "adad"
	Len   = ad.Len + splice.AddrLen
)

// Ad stores a specification for the fee rate and the service ports of a set of
// services being offered at a relay
//
// These must be a well known port to match with a type of service, eg 80 for
// web, 53 for DNS, 8333 for bitcoin p2p, 8334 for bitcoin JSONRPC... For
// simplicity.
type Ad struct {

	// Embed ad.Ad for the common fields
	ad.Ad

	// Addresses that the peer can be reached on.
	Addresses []*netip.AddrPort
}

var _ codec.Codec = &Ad{}

// New creates a new addresses.Ad.
func New(id nonce.ID, key *crypto.Prv, addrs []*netip.AddrPort,
	expiry time.Time) (addrAd *Ad) {

	k := crypto.DerivePub(key)
	addrAd = &Ad{
		Ad: ad.Ad{
			ID:     id,
			Key:    k,
			Expiry: expiry,
		},
		Addresses: addrs,
	}
	log.T.S("address ad", addrAd)
	if e := addrAd.Sign(key); fails(e) {
		return
	}
	log.T.S("signed", addrAd)
	return
}

// Decode a splice.Splice's next bytes into an Ad.
func (x *Ad) Decode(s *splice.Splice) (e error) {
	var i, count uint16
	s.ReadID(&x.ID).
		ReadPubkey(&x.Key).
		ReadUint16(&count)
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

// Unwrap returns nothing because there isn't an onion inside an Ad.
func (x *Ad) Unwrap() interface{} { return nil }

// Len returns the length of bytes required to encode the Ad, based on the number
// of Addresses inside it.
func (x *Ad) Len() int {
	return ad.Len + len(x.Addresses)*(1+Len) + slice.Uint16Len
}

// Magic bytes that identify this message
func (x *Ad) Magic() string { return Magic }

// Sign the encoded form of the bytes in order to authorise it.
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

// Validate checks that the signature matches the public key.
func (x *Ad) Validate() (valid bool) {
	s := splice.New(x.Len())
	x.SpliceNoSig(s)
	return x.Sig.MatchesPubkey(s.GetUntilCursor(), x.Key) &&
		x.Expiry.After(time.Now())
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

// Splice is a function that serializes the parts of an Ad.
func Splice(s *splice.Splice, id nonce.ID, key *crypto.Pub,
	addrs []*netip.AddrPort, expiry time.Time) {

	s.Magic(Magic).
		ID(id).
		Pubkey(key).
		Uint16(uint16(len(addrs)))
	for i := range addrs {
		s.AddrPort(addrs[i])
	}
	s.Time(expiry)
}

func init() { reg.Register(Magic, Gen) }

// Gen is a factory function to generate an Ad.
func Gen() codec.Codec { return &Ad{} }
