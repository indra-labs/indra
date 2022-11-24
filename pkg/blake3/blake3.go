// Package blake3 provides a simple interface for single and double SHA256
// hashes, used with secp256k1 signatures, message digest checksums, cloaked
// public key "addresses" and so on.
package blake3

import (
	"fmt"
	"unsafe"

	"github.com/zeebo/blake3"
)

const (
	Len          = 32
	errLengthStr = "invalid hash length of %d bytes, must be %d"
)

// Hash is just a byte type with a nice length validation and zeroing function.
type Hash []byte

func ErrorLength(l int) error { return fmt.Errorf(errLengthStr, l, Len) }

// New creates a correctly sized slice for a Hash.
func New() Hash { return make(Hash, Len) }

// Double runs a standard double SHA256 hash and does all the slicing for you.
func Double(b []byte) Hash { return Single(Single(b)) }

// Single runs a standard SHA256 hash and does all the slicing for you.
func Single(b []byte) Hash { return FromHash(blake3.Sum256(b)) }

// Valid checks the hash value is the correct length.
func (h Hash) Valid() (e error) {
	if len(h) != Len {
		e = ErrorLength(len(h))
	}
	return
}

func (h Hash) Equals(h2 Hash) bool {
	// Ensure lengths are correct.
	if len(h) == Len && len(h2) == Len {
		return *(*string)(unsafe.Pointer(&h)) ==
			*(*string)(unsafe.Pointer(&h2))
	}
	return false
}

// this is made as a string to be immutable. It can be changed with unsafe ofc.
var zero = string(New())

// Zero out the values in the hash. Hashes can be used as secrets.
func (h Hash) Zero() { copy(h, zero) }

func FromHash(b [32]byte) []byte { return b[:] }
