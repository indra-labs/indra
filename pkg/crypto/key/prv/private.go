// Package prv is a wrapper around secp256k1 library from the Decred project to
// handle, generate and serialise secp256k1 private keys.
package prv

import (
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	
	"git-indra.lan/indra-labs/indra"
	"git-indra.lan/indra-labs/indra/pkg/b32/based32"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

const (
	KeyLen = secp256k1.PrivKeyBytesLen
)

// Key is a private key.
type Key secp256k1.PrivateKey
type Bytes [KeyLen]byte

// GenerateKey a private key.
func GenerateKey() (prv *Key, e error) {
	var p *secp256k1.PrivateKey
	if p, e = secp256k1.GeneratePrivateKey(); check(e) {
		return
	}
	return (*Key)(p), e
}

func (k *Key) ToBase32() (s string) {
	b := k.ToBytes()
	var e error
	if s, e = based32.Codec.Encode(b[:]); check(e) {
	}
	ss := []byte(s[1:])
	return string(ss)
}

func FromBase32(s string) (k *Key, e error) {
	ss := []byte(s)
	var b slice.Bytes
	b, e = based32.Codec.Decode("a" + string(ss))
	k = PrivkeyFromBytes(b)
	return
}

// PrivkeyFromBytes converts a byte slice into a private key.
func PrivkeyFromBytes(b []byte) *Key {
	return (*Key)(secp256k1.PrivKeyFromBytes(b))
}

// Zero out a private key to prevent key scraping from memory.
func (prv *Key) Zero() { (*secp256k1.PrivateKey)(prv).Zero() }

// ToBytes returns the Bytes serialized form. It zeroes the original bytes.
func (prv *Key) ToBytes() (b Bytes) {
	br := (*secp256k1.PrivateKey)(prv).Serialize()
	copy(b[:], br[:KeyLen])
	// zero the original
	copy(br, zero())
	return
}

func zero() []byte {
	z := Bytes{}
	return z[:]
}

// Zero zeroes out a private key in serial form.
func (pb Bytes) Zero() { copy(pb[:], zero()) }
