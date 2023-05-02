package crypto

import (
	"testing"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/util/tests"
)

// this just really demonstrates how the keys generated are not linkable.
func TestKeySet_Next(t *testing.T) {
	var counter int
	const nRounds = 10000
	for rounds := 0; rounds < nRounds; rounds++ {
		key, ks, e := NewSigner()
		if fails(e) {
			t.FailNow()
		}
		var hx PubBytes
		if hx = DerivePub(key).ToBytes(); fails(e) {
			t.Error(e)
		}
		oddness := hx[0]
		// for i := 0; i < 100; i++ {
		key = ks.Next()
		if hx = DerivePub(key).ToBytes(); hx[0] == oddness {
			counter++
		}
		// }
	}
	if counter == nRounds || counter == 0 {
		t.Error("all keys same oddness", counter)
	}
}

func BenchmarkKeySet_Next(b *testing.B) {
	_, ks, e := NewSigner()
	if fails(e) {
		b.FailNow()
	}
	for n := 0; n < b.N; n++ {
		_ = ks.Next()
	}
}

func BenchmarkKeySet_Next_Derive(b *testing.B) {
	_, ks, e := NewSigner()
	if fails(e) {
		b.FailNow()
	}
	for n := 0; n < b.N; n++ {
		k := ks.Next()
		DerivePub(k)
	}
}

func BenchmarkKeySet_Next_Sign(b *testing.B) {
	_, ks, e := NewSigner()
	if fails(e) {
		b.FailNow()
	}
	var msg []byte
	const msgLen = 1382 - 4 - SigLen
	msg, _, e = tests.GenMessage(msgLen, "herpderp")
	for n := 0; n < b.N; n++ {
		k := ks.Next()
		hash := sha256.Single(msg)
		if _, e = Sign(k, hash); fails(e) {
			b.Error("failed to sign")
		}
	}
}
