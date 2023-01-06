// Package pub is a wrapper around secp256k1 library from the Decred project to
// handle generate and serialise secp256k1 public keys, including deriving them
// from private keys.
package pub

import (
	"encoding/hex"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/indra-labs/indra"
	"github.com/indra-labs/indra/pkg/key/prv"
	log2 "github.com/indra-labs/indra/pkg/log"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

const (
	// KeyLen is the length of the serialized key. It is an ECDSA compressed
	// key.
	KeyLen = secp256k1.PubKeyBytesLenCompressed
)

type (
	// Key is a public key.
	Key secp256k1.PublicKey
	// Bytes is the serialised form of a public key.
	Bytes [KeyLen]byte
)

// Derive generates a public key from the prv.Key.
func Derive(prv *prv.Key) *Key {
	return (*Key)((*secp256k1.PrivateKey)(prv).PubKey())
}

// FromBytes converts a byte slice into a public key, if it is valid and on the
// secp256k1 elliptic curve.
func FromBytes(b []byte) (pub *Key, e error) {
	var p *secp256k1.PublicKey
	if p, e = secp256k1.ParsePubKey(b); check(e) {
		return
	}
	pub = (*Key)(p)
	return
}

// ToBytes returns the compressed 33 byte form of the pubkey as used in wire and
// storage forms.
func (pub *Key) ToBytes() (p Bytes) {
	b := (*secp256k1.PublicKey)(pub).SerializeCompressed()
	copy(p[:], b)
	return
}

func (pub *Key) ToHex() (s string, e error) {
	b := pub.ToBytes()
	s = hex.EncodeToString(b[:])
	return
}

func (pb Bytes) Equals(qb Bytes) bool { return pb == qb }

func (pub *Key) ToPublicKey() *secp256k1.PublicKey {
	return (*secp256k1.PublicKey)(pub)
}

// Equals returns true if two public keys are the same.
func (pub *Key) Equals(pub2 *Key) bool {
	return pub.ToPublicKey().IsEqual(pub2.ToPublicKey())
}
