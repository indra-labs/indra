package sha256

import "crypto/sha256"

// Double runs a standard double SHA256 hash and does all the slicing for you.
func Double(b []byte) []byte {
	return Hash(Hash(b))
}

// Hash runs a standard SHA256 hash and does all the slicing for you.
func Hash(b []byte) []byte {
	h := sha256.Sum256(b)
	return h[:]
}
