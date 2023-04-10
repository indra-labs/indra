package crypto

import (
	"crypto/rand"
	"testing"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
)

// this just really demonstrates how the keys generated are not linkable.
func TestKeySet_Next(t *testing.T) {
	for rounds := 0; rounds < 1000; rounds++ {
		key, ks, e := NewSigner()
		if check(e) {
			t.FailNow()
		}
		var hx PubBytes
		if hx = DerivePub(key).ToBytes(); check(e) {
			t.Error(e)
		}
		oddness := hx[0]
		var counter int
		for i := 0; i < 100; i++ {
			key = ks.Next()
			if hx = DerivePub(key).ToBytes(); hx[0] == oddness {
				counter++
			}
		}
		if counter == 100 || counter == 0 {
			t.Error("all keys same oddness")
		}
	}
}

func BenchmarkKeySet_Next(b *testing.B) {
	_, ks, e := NewSigner()
	if check(e) {
		b.FailNow()
	}
	for n := 0; n < b.N; n++ {
		_ = ks.Next()
	}
}

func BenchmarkKeySet_Next_Derive(b *testing.B) {
	_, ks, e := NewSigner()
	if check(e) {
		b.FailNow()
	}
	for n := 0; n < b.N; n++ {
		k := ks.Next()
		DerivePub(k)
	}
}

func GenerateTestMessage(msgSize int) (msg []byte, hash sha256.Hash, e error) {
	msg = make([]byte, msgSize)
	var n int
	if n, e = rand.Read(msg); check(e) && n != msgSize {
		return
	}
	copy(msg, "payload")
	hash = sha256.Single(msg)
	return
}

func BenchmarkKeySet_Next_Sign(b *testing.B) {
	_, ks, e := NewSigner()
	if check(e) {
		b.FailNow()
	}
	var msg []byte
	const msgLen = 1382 - 4 - SigLen
	msg, _, e = GenerateTestMessage(msgLen)
	for n := 0; n < b.N; n++ {
		k := ks.Next()
		hash := sha256.Single(msg)
		if _, e = Sign(k, hash); check(e) {
			b.Error("failed to sign")
		}
	}
}
