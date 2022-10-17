package sha256

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestHash(t *testing.T) {
	// this generates the rfc6979ExtraDataV0 value in the SHA256 based
	// Schnorr signature algorithm forked from Decred's blake256 based one.
	hash := sha256.Sum256([]byte("EC-Schnorr-INDRAv0"))
	hexed := hex.EncodeToString(hash[:])
	var hexbytes []string
	for i := 0; i < len(hexed); i += 2 {
		hexbytes = append(hexbytes, "0x"+hexed[i:i+2]+",")
	}
	t.Log(hexbytes)
}
