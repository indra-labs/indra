package pub

import (
	"encoding/hex"
	"fmt"
	"unsafe"
)

const (
	PrintLen = 6
)

type (
	Print []byte
)

// GetPrints creates a slice of fingerprints from a set of public keys.
func GetPrints(keys ...*Key) (fps []Print) {
	for i := range keys {
		fps = append(fps, keys[i].ToBytes().Fingerprint())
	}
	return
}

func (p Print) String() string {
	return hex.EncodeToString(p)
}

func (p Print) Equals(q Print) (e error) {
	// Ensure lengths are correct.
	if len(p) == PrintLen && len(q) == PrintLen {
		// Convert to fixed arrays unsafely, but now safe in fact, and
		// we get access to the builtin comparison operator and don't
		// dice with the immutability of strings.
		same := *(*string)(unsafe.Pointer(&p)) ==
			*(*string)(unsafe.Pointer(&q))
		if !same {
			e = fmt.Errorf("pubkey print %s != %s", p, q)
		}
	} else {
		e = fmt.Errorf("prints not the same length, %d != %d",
			len(p), len(q))
	}
	return
}
