// Package nonce provides a simple interface for generating standard AES
// encryption nonces that give strong cryptographic entropy to message
// encryption.
package nonce

import (
	"crypto/aes"
	"crypto/rand"
	"github.com/indra-labs/indra"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
)

const IVLen = aes.BlockSize

var (
	log   = log2.GetLogger(indra.PathBase)
	fails = log.E.Chk
)

type IV [IVLen]byte

// New reads a nonce from a cryptographically secure random number source
func New() (n IV) {
	if c, e := rand.Read(n[:]); fails(e) && c != IDLen {
	}
	return
}
