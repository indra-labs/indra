package sha256

import (
	"testing"
)

func TestHash_Equals(t *testing.T) {
	var h1, h2 Hash
	if h1 != h2 {
		t.FailNow()
	}
	h2[0] = 1
	if h1 == h2 {
		t.FailNow()
	}
}
