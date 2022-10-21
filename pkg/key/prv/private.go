package prv

import (
	"github.com/Indra-Labs/indra"
	log2 "github.com/cybriq/proc/pkg/log"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
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
type Bytes []byte

// GenerateKey a private key.
func GenerateKey() (prv *Key, e error) {
	var p *secp256k1.PrivateKey
	if p, e = secp256k1.GeneratePrivateKey(); check(e) {
		return
	}
	return (*Key)(p), e
}

// PrivkeyFromBytes converts a byte slice into a private key.
func PrivkeyFromBytes(b Bytes) *Key {
	return (*Key)(secp256k1.PrivKeyFromBytes(b))
}

// Zero out a private key to prevent key scraping from memory.
func (prv *Key) Zero() { (*secp256k1.PrivateKey)(prv).Zero() }

// ToBytes returns the Bytes serialized form.
func (prv *Key) ToBytes() (b Bytes) {
	return (*secp256k1.PrivateKey)(prv).Serialize()[:KeyLen]
}

// this is made as a string to be immutable. It can be changed with unsafe ofc.
var zero = string(make([]byte, KeyLen))

// Zero zeroes out a private key in serial form.
func (pb Bytes) Zero() { copy(pb, zero[:]) }
