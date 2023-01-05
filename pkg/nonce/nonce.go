// Package nonce provides a simple interface for generating standard AES
// encryption nonces that give strong cryptographic entropy to message
// encryption.
package nonce

import (
	"crypto/aes"
	"crypto/rand"

	"github.com/Indra-Labs/indra"
	log2 "github.com/Indra-Labs/indra/pkg/log"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

const IVLen = aes.BlockSize

type IV [IVLen]byte

// New reads a nonce from a cryptographically secure random number source
func New() (n IV) {
	if _, e := rand.Read(n[:]); check(e) {
	}
	return
}
