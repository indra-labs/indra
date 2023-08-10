package crypto

import (
	"git.indra-labs.org/dev/ind"
	"testing"

	log2 "git.indra-labs.org/dev/ind/pkg/proc/log"
)

func TestBase32(t *testing.T) {
	if indra.CI == "false" {
		log2.SetLogLevel(log2.Debug)
	}
	for i := 0; i < 10000; i++ {
		var k *Prv
		var e error
		if k, e = GeneratePrvKey(); fails(e) {
			t.Error(e)
			t.FailNow()
		}
		p := DerivePub(k)
		b32 := p.ToBased32()
		pr32 := k.ToBased32()
		//log.D.F("pub key: %d %s priv key: %d %s\n", len(b32), b32, len(pr32), pr32)
		var kk *Pub
		if kk, e = PubFromBased32(b32); fails(e) {
			t.Error(e)
			t.FailNow()
		}
		if b32 != kk.ToBased32() {
			t.Error("failed to decode public key")
			t.FailNow()
		}
		var pk *Prv
		if pk, e = PrvFromBased32(pr32); fails(e) {
			t.Error(e)
			t.FailNow()
		}
		if pr32 != pk.ToBased32() {
			t.Error("failed to decode private key")
			t.FailNow()
		}
	}
}
