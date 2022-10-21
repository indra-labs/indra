package sha256

import (
	"fmt"

	"github.com/Indra-Labs/indra/pkg/slice"
	"github.com/minio/sha256-simd"
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
func Double(b []byte) []byte { return Single(Single(b)) }

// Single runs a standard SHA256 hash and does all the slicing for you.
func Single(b []byte) []byte { return slice.FromHash(sha256.Sum256(b)) }

// Valid checks the hash value is the correct length.
func (h Hash) Valid() (e error) {
	if len(h) != Len {
		e = ErrorLength(len(h))
	}
	return
}

// this is made as a string to be immutable. It can be changed with unsafe ofc.
var zero = string(New())

// Zero out the values in the hash. Hashes can be used as secrets.
func (h Hash) Zero() { copy(h, zero) }
