package ngin

import (
	"testing"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

func TestOnionSkins_SimpleCrypt(t *testing.T) {
	log2.SetLogLevel(log2.Trace)
	var e error
	n := nonce.NewID()
	n1 := nonce.New()
	prv1, prv2 := GetTwoPrvKeys(t)
	pub1, pub2 := pub.Derive(prv1), pub.Derive(prv2)
	on := Skins{}.
		Crypt(pub1, pub2, prv2, n1, 0).
		Confirmation(n, 0).
		Assemble()
	s := Encode(on)
	s.SetCursor(0)
	log.D.S("encoded, encrypted onion:\n", s.GetRange(-1, -1).ToBytes())
	var oncr Onion
	if oncr = Recognise(s, slice.GenerateRandomAddrPortIPv6()); oncr == nil {
		t.Error("did not unwrap")
		t.FailNow()
	}
	if e = oncr.Decode(s); check(e) {
		t.Error("did not decode")
		t.FailNow()
	}
	oncr.(*Crypt).Decrypt(prv1, s)
	var oncn Onion
	if oncn = Recognise(s, slice.GenerateRandomAddrPortIPv6()); oncn == nil {
		t.Error("did not unwrap")
		t.FailNow()
	}
	if e = oncn.Decode(s); check(e) {
		t.Error("did not decode")
		t.FailNow()
	}
	if cn, ok := oncn.(*Confirmation); !ok {
		t.Error("did not get expected confirmation")
		t.FailNow()
	} else {
		if cn.ID != n {
			t.Error("did not get expected confirmation ID")
			t.FailNow()
		}
	}
}
