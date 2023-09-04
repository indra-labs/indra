package crypto

import (
	rand2 "crypto/rand"
	"git.indra-labs.org/dev/ind/pkg/util/ci"
	"testing"
	
	"git.indra-labs.org/dev/ind/pkg/crypto/sha256"
)

func TestFromBased32(t *testing.T) {
	ci.TraceIfNot()
	var rBytes sha256.Hash
	var n int
	var e error
	if n, e = rand2.Read(rBytes[:]); n != 8 && fails(e) {
		t.FailNow()
	}
	var pr *Prv
	if pr, e = GeneratePrvKey(); fails(e) {
		t.FailNow()
	}
	for i := 0; i < 1<<16; i++ {
		var s SigBytes
		if s, e = Sign(pr, rBytes); fails(e) {
			t.FailNow()
		}
		// fmt.Println("sig", s)
		var sb SigBytes
		if sb, e = SigFromBased32(s.String()); fails(e) {
			t.FailNow()
		}
		if s != sb {
			t.Error("sig mismatch")
			t.FailNow()
		}
		rBytes = sha256.Single(rBytes[:])
	}
}
