// Package pub is a wrapper around secp256k1 library from the Decred project to
// handle generate and serialise secp256k1 public keys, including deriving them
// from private keys.
package pub

import (
	"encoding/base32"
	"encoding/hex"
	
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/gookit/color"
	
	"git-indra.lan/indra-labs/indra"
	"git-indra.lan/indra-labs/indra/pkg/b32/based32"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
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

var enc = base32.NewEncoding(Charset).EncodeToString

const Charset = "abcdefghijklmnopqrstuvwxyz234679"

type (
	// Key is a public key.
	Key secp256k1.PublicKey
	// Bytes is the serialised form of a public key.
	Bytes [KeyLen]byte
)

func (pb Bytes) String() (s string) {
	var e error
	if s, e = based32.Codec.Encode(pb[:]); check(e) {
	}
	ss := []byte(s)
	// Reverse text order to get all starting ciphers.
	for i := 0; i < len(s)/2; i++ {
		ss[i], ss[len(s)-i-1] = ss[len(s)-i-1], ss[i]
	}
	return color.LightGreen.Sprint(string(ss))
}

func (k *Key) String() (s string) {
	return k.ToBase32()
}

// Derive generates a public key from the prv.Key.
func Derive(prv *prv.Key) *Key {
	if prv == nil {
		return nil
	}
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
func (k *Key) ToBytes() (p Bytes) {
	b := (*secp256k1.PublicKey)(k).SerializeCompressed()
	copy(p[:], b)
	return
}

func (k *Key) ToHex() (s string, e error) {
	b := k.ToBytes()
	s = hex.EncodeToString(b[:])
	return
}

func (k *Key) ToBase32() (s string) {
	b := k.ToBytes()
	var e error
	if s, e = based32.Codec.Encode(b[:]); check(e) {
	}
	ss := []byte(s)
	// Reverse text order to get all starting ciphers.
	for i := 0; i < len(s)/2; i++ {
		ss[i], ss[len(s)-i-1] = ss[len(s)-i-1], ss[i]
	}
	return color.LightGreen.Sprint(string(ss))
}

func (k *Key) ToBase32Abbreviated() (s string) {
	b := k.ToBytes()
	var e error
	if s, e = based32.Codec.Encode(b[:]); check(e) {
	}
	ss := []byte(s)
	// Reverse text order to get all starting ciphers.
	for i := 0; i < len(s)/2; i++ {
		ss[i], ss[len(s)-i-1] = ss[len(s)-i-1], ss[i]
	}
	ss = append(ss[:13], append([]byte("..."), ss[len(ss)-8:]...)...)
	return color.LightGreen.Sprint(string(ss))
}

func FromBase32(s string) (k *Key, e error) {
	ss := []byte(s)
	// Reverse text order to get all starting ciphers.
	for i := 0; i < len(s)/2; i++ {
		ss[i], ss[len(s)-i-1] = ss[len(s)-i-1], ss[i]
	}
	var b slice.Bytes
	b, e = based32.Codec.Decode(string(ss))
	return FromBytes(b)
}

func (pb Bytes) Equals(qb Bytes) bool { return pb == qb }

func (k *Key) ToPublicKey() *secp256k1.PublicKey {
	return (*secp256k1.PublicKey)(k)
}

// Equals returns true if two public keys are the same.
func (k *Key) Equals(pub2 *Key) bool {
	return k.ToPublicKey().IsEqual(pub2.ToPublicKey())
}
