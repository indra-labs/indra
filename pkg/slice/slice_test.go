package slice

import (
	"testing"
)

func TestSize24(t *testing.T) {
	n := 1<<24 - 1
	log.I.Ln(n)
	u := NewUint24()
	EncodeUint24(u, n)
	u2 := DecodeUint24(u)
	if n != u2 {
		t.Error("failed to encode/decode")
	}
}
