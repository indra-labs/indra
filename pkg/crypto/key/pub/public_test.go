package pub

import (
	"testing"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
)

func TestBase32(t *testing.T) {
	log2.SetLogLevel(log2.Trace)
	for i := 0; i < 10000; i++ {
		var k *prv.Key
		var e error
		if k, e = prv.GenerateKey(); check(e) {
			t.Error(e)
			t.FailNow()
		}
		p := Derive(k)
		b32 := p.ToBase32()
		pr32 := k.ToBase32()
		log.I.Ln("\n", len(b32), b32, len(pr32), pr32)
		var kk *Key
		kk, e = FromBase32(b32)
		if b32 != kk.ToBase32() {
			t.Error(e)
			t.FailNow()
		}
	}
}
