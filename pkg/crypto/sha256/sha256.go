// Package sha256 provides a simple interface for single and double SHA256
// hashes, used with secp256k1 signatures, message digest checksums, cloaked
// public key "addresses" and so on.
package sha256

import (
	"encoding/base32"
	"encoding/hex"
	"github.com/indra-labs/indra/pkg/constant"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/minio/sha256-simd"
)

const Len = 32

var (
	// enc is a raw base32 encoder as 256 bit hashes have a consistent set of
	// extraneous characters after 52 digits from padding and do not need check
	// bytes as they are compact large numbers for logs and message digests for
	// other things.
	enc   = base32.NewEncoding(constant.Based32Ciphers).EncodeToString
	log   = log2.GetLogger()
	fails = log.E.Chk
)

// String returns the hex encoded form of a SHA256 hash.
func (h Hash) String() string {
	return hex.EncodeToString(h[:])
}

// Based32String returns the Basde32 encoded form of a SHA256 hash.
func (h Hash) Based32String() string {
	return enc(h[:])[:52]
}

// Hash is just a 256-bit hash.
type Hash [32]byte

// Double runs a standard double SHA256 hash and does all the slicing for you.
func Double(b []byte) Hash {
	h := Single(b)
	return Single(h[:])
}

// New creates a correctly sized slice for a Hash.
func New() Hash { return Hash{} }

// Single runs a standard SHA256 hash.
func Single(b []byte) Hash { return sha256.Sum256(b) }

// Zero copies a cleanly initialised empty slice over top of the provided Hash.
func Zero(h Hash) { copy(h[:], zero()) }

// Zero out the values in the hash. Hashes can be used as secrets.
func (h Hash) Zero() { Zero(h) }

func zero() []byte { return make([]byte, Len) }
