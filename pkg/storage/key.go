package storage

import (
	"crypto/rand"
	"github.com/btcsuite/btcd/btcutil/base58"
)

type Key [32]byte

func (k Key) Bytes() []byte {
	return k[:]
}

func (k Key) Encode() string {
	return base58.Encode(k[:])
}

func (k Key) Decode(key string) {
	base58.Decode(key)
}

func KeyGen() (Key, error) {

	var err error
	var sk [32]byte
	var key Key

	_, err = rand.Read(sk[:])

	sk[0] &= 248
	sk[31] = (sk[31] & 127) | 64

	key = sk

	return key, err
}
