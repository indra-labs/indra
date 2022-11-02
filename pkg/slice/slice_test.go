package slice

import (
	"testing"
)

func TestSize24(t *testing.T) {
	n := 1<<24 - 1
	log.I.Ln(n)
	u := NewUint24()
	EncodeUint24(u, n)
	log.I.S(u)
	log.I.S(DecodeUint24(u))
}
