package pub

import (
	"testing"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
)

func TestBase32(t *testing.T) {
	for i := 0; i < 1000; i++ {
		var k *prv.Key
		var e error
		if k, e = prv.GenerateKey(); check(e) {
			t.Error(e)
			t.FailNow()
		}
		p := Derive(k)
		b32 := p.ToBase32()
		log.I.Ln(b32)
		var kk *Key
		kk, e = FromBase32(b32)
		if b32 != kk.ToBase32() {
			t.Error(e)
			t.FailNow()
		}
	}
}
