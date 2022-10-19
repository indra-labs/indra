package hasj

import (
	"fmt"
)

const HashLen = 32

// Hash is just a byte type with a nice length validation and zeroing function.
type Hash []byte

// Valid checks the
func (h Hash) Valid() error {
	if len(h) == HashLen {
		return nil
	}
	return fmt.Errorf("invalid hash length of %d bytes, must be %d",
		len(h), HashLen)
}

func (h Hash) Zero() {
	for i := range h {
		h[i] = 0
	}
}
