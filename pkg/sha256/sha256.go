// Package sha256 provides a simple interface for single and double SHA256
// hashes, used with secp256k1 signatures, message digest checksums, cloaked
// public key "addresses" and so on.
package sha256

import (
	"github.com/indra-labs/indra"
	log2 "github.com/indra-labs/indra/pkg/log"
	"github.com/minio/sha256-simd"
)

const (
	Len = 32
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

// Hash is just a 256-bit hash.
type Hash [32]byte

// New creates a correctly sized slice for a Hash.
func New() Hash { return Hash{} }

// Double runs a standard double SHA256 hash and does all the slicing for you.
func Double(b []byte) Hash {
	h := Single(b)
	return Single(h[:])
}

// Single runs a standard SHA256 hash and does all the slicing for you.
func Single(b []byte) Hash { return sha256.Sum256(b) }

func zero() []byte { return make([]byte, Len) }

// Zero copies a cleanly initialised empty slice over top of the provided Hash.
func Zero(h Hash) { copy(h[:], zero()) }

// Zero out the values in the hash. Hashes can be used as secrets.
func (h Hash) Zero() { Zero(h) }
