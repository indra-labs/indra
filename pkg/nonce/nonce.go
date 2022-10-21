package nonce

import (
	"crypto/aes"
	"crypto/rand"

	"github.com/Indra-Labs/indra"
	log2 "github.com/cybriq/proc/pkg/log"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

const Size = aes.BlockSize

type IV []byte

// Get reads a nonce from a cryptographically secure random number source
func Get() (n IV) {
	n = IV{}
	if _, e := rand.Read(n[:]); check(e) {
	}
	return
}
