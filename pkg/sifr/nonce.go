package sifr

import (
	"crypto/aes"
	"crypto/rand"
)

const NonceSize = aes.BlockSize

type Nonce [NonceSize]byte

// GetNonce reads from a cryptographically secure random number source
func GetNonce() (nonce *Nonce) {
	nonce = &Nonce{}
	if _, e := rand.Read(nonce[:]); log.E.Chk(e) {
	}
	return
}
