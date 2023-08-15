// Package addresses defines the message format that provides the network multi-address of a peer with a given public identity key.
package addresses

import (
	"git.indra-labs.org/dev/ind/pkg/cfg"
	"git.indra-labs.org/dev/ind/pkg/codec"
	"git.indra-labs.org/dev/ind/pkg/codec/ad"
	"git.indra-labs.org/dev/ind/pkg/codec/reg"
	"git.indra-labs.org/dev/ind/pkg/crypto"
	"git.indra-labs.org/dev/ind/pkg/crypto/nonce"
	log2 "git.indra-labs.org/dev/ind/pkg/proc/log"
	"git.indra-labs.org/dev/ind/pkg/util/slice"
	"git.indra-labs.org/dev/ind/pkg/util/splice"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"time"
)

var (
	log   = log2.GetLogger()
	fails = log.E.Chk
)

const (
	Magic = "adad"
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
	Addresses []multiaddr.Multiaddr
}

var _ codec.Codec = &Ad{}

// New creates a new addresses.Ad.
func New(id nonce.ID, key *crypto.Prv, addrs []multiaddr.Multiaddr,
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

func (x *Ad) PubKey() (key *crypto.Pub) { return x.Key }
func (x *Ad) Fingerprint() (pf string)  { return x.Key.Fingerprint() }
func (x *Ad) Expired() (is bool)        { return x.Expiry.Before(time.Now()) }

func (x *Ad) GetID() (id peer.ID, e error) {
	return peer.IDFromPublicKey(x.Key)
}

// Decode a splice.Splice's next bytes into an Ad.
func (x *Ad) Decode(s *splice.Splice) (e error) {
	var i, count uint16
	s.ReadID(&x.ID).
		ReadPubkey(&x.Key).
		ReadUint16(&count)
	x.Addresses = make([]multiaddr.Multiaddr, count)
	for ; i < count; i++ {
		var addy multiaddr.Multiaddr
		s.ReadMultiaddr(&addy)
		x.Addresses[i] = addy
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

// Unwrap returns nothing because there isn't an onion inside an Ad.
func (x *Ad) Unwrap() interface{} { return nil }

// Len returns the length of bytes required to encode the Ad, based on the number
// of Addresses inside it.
func (x *Ad) Len() int {
	
	codec.MustNotBeNil(x)
	
	l := ad.Len + slice.Uint16Len
	// Generate the addresses to get their data length:
	// for _, v := range x.Addresses {
	// var b []byte
	// var e error
	// b, e = multi.AddrToBytes(v, cfg.GetCurrentDefaultPort())
	// if fails(e) {
	// 	panic(e)
	// }
	// log.D.S("bytes", b)
	// l += 21 // len(b) + 1
	// }
	l += 21 * len(x.Addresses)
	// log.D.Ln(l)
	return l
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

// SpliceNoSig splices until the signature.
func (x *Ad) SpliceNoSig(s *splice.Splice) {
	Splice(s, x.ID, x.Key, x.Addresses, x.Expiry)
}

// Splice is a function that serializes the parts of an Ad.
func Splice(s *splice.Splice, id nonce.ID, key *crypto.Pub,
	addrs []multiaddr.Multiaddr, expiry time.Time) {
	
	// log.D.Ln("spliced", s.GetPos())
	s.Magic(Magic)
	// log.D.Ln("spliced", s.GetPos(), "magic")
	s.ID(id)
	// log.D.Ln("spliced", s.GetPos(), "ID")
	s.Pubkey(key)
	// log.D.Ln("spliced", s.GetPos(), "pubkey")
	s.Uint16(uint16(len(addrs)))
	// log.D.Ln("spliced", s.GetPos(), "addrlen")
	for i := range addrs {
		s.Multiaddr(addrs[i], cfg.GetCurrentDefaultPort())
		// log.D.Ln("addresses", s.GetPos(), "addr", i)
	}
	s.Time(expiry)
	// log.D.Ln("spliced", s.GetPos(), "bytes")
}

func init() { reg.Register(Magic, Gen) }

// Gen is a factory function to generate an Ad.
func Gen() codec.Codec { return &Ad{} }
