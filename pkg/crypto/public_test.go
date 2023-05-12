package crypto

import (
	"testing"
	
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
)

func TestBase32(t *testing.T) {
	log2.SetLogLevel(log2.Debug)
	for i := 0; i < 10000; i++ {
		var k *Prv
		var e error
		if k, e = GeneratePrvKey(); fails(e) {
			t.Error(e)
			t.FailNow()
		}
		p := DerivePub(k)
		b32 := p.ToBase32()
		pr32 := k.ToBase32()
		log.I.Ln("\n", len(b32), b32, len(pr32), pr32)
		var kk *Pub
		kk, e = PubFromBase32(b32)
		if b32 != kk.ToBase32() {
			t.Error(e)
			t.FailNow()
		}
	}
}
