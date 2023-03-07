package engine

import (
	"testing"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/signer"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

func TestOnionSkins_HiddenService(t *testing.T) {
	var e error
	n3 := Gen3Nonces()
	id := nonce.NewID()
	pr, ks, _ := signer.New()
	in := NewIntro(pr, slice.GenerateRandomAddrPortIPv6())
	var prvs [3]*prv.Key
	for i := range prvs {
		prvs[i] = ks.Next()
	}
	var pubs [3]*pub.Key
	for i := range pubs {
		pubs[i] = pub.Derive(prvs[i])
	}
	on1 := Skins{}.
		HiddenService(id, in, prvs, pubs, n3)
	on1 = append(on1, &Tmpl{})
	on := on1.Assemble()
	s := Encode(on)
	s.SetCursor(0)
	var onc Onion
	if onc = Recognise(s); onc == nil {
		t.Error("did not unwrap")
		t.FailNow()
	}
	if e = onc.Decode(s); check(e) {
		t.Error("did not decode")
		t.FailNow()
	}
	var ex *HiddenService
	var ok bool
	if ex, ok = onc.(*HiddenService); !ok {
		t.Error("did not unwrap expected type")
		t.FailNow()
	}
	if ex.ID != id {
		t.Error("ID did not decode correctly")
		t.FailNow()
	}
	for i := range ex.Ciphers {
		if ex.Ciphers[i] != on.(*HiddenService).Ciphers[i] {
			t.Errorf("cipher %d did not unwrap correctly", i)
			t.FailNow()
		}
	}
	for i := range ex.Nonces {
		if ex.Nonces[i] != n3[i] {
			t.Errorf("nonce %d did not unwrap correctly", i)
			t.FailNow()
		}
	}
	if !ex.Intro.Key.Equals(in.Key) {
		t.Errorf("key did not decode correctly")
		t.FailNow()
	}
	if ex.AddrPort.String() != in.AddrPort.String() {
		t.Errorf("addrport did not decode correctly")
		t.FailNow()
	}
	if string(ex.Intro.Sig[:]) != string(in.Sig[:]) {
		t.Errorf("signature did not decode correctly")
		t.FailNow()
	}
	if !ex.Intro.Validate() {
		t.Errorf("received intro did not validate")
		t.FailNow()
	}
}
