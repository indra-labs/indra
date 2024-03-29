// Package nonce provides a simple interface for generating standard AES
// encryption nonces that give strong cryptographic entropy to message
// encryption, as well as 8 byte (64 bit) random private identifiers for
// references between types.
package nonce

import (
	"crypto/aes"
	"crypto/rand"
	log2 "git.indra-labs.org/dev/ind/pkg/proc/log"
)

// IVLen is the length of Initialization Vectors used in Indra.
const IVLen = aes.BlockSize

var (
	log   = log2.GetLogger()
	fails = log.E.Chk
)

// IV is an Initialization Vector for AES-CTR encryption used in Indra.
type IV [IVLen]byte

// New reads a nonce from a cryptographically secure random number source.
func New() (n IV) {
	if c, e := rand.Read(n[:]); fails(e) && c != IDLen {
	}
	return
}
