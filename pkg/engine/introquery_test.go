package engine

import (
	"testing"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
)

func TestOnionSkins_IntroQuery(t *testing.T) {
	var e error
	prvs, pubs := GetCipherSet(t)
	ciphers := GenCiphers(prvs, pubs)
	prv1, _ := GetTwoPrvKeys(t)
	pub1 := pub.Derive(prv1)
	n3 := Gen3Nonces()
	ep := &ExitPoint{
		Routing: &Routing{
			Sessions: [3]*SessionData{},
			Keys:     prvs,
			Nonces:   n3,
		},
		ReturnPubs: pubs,
	}
	on := Skins{}.
		IntroQuery(pub.Derive(prv1), ep).
		Tmpl().Assemble()
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
	var ex *IntroQuery
	var ok bool
	if ex, ok = onc.(*IntroQuery); !ok {
		t.Error("did not unwrap expected type")
		t.FailNow()
	}
	for i := range ex.Ciphers {
		if ex.Ciphers[i] != ciphers[i] {
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
	if !ex.Key.Equals(pub1) {
		t.Error("Key did not decode correctly")
		t.FailNow()
	}
}
