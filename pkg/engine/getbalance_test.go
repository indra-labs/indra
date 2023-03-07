package engine

import (
	"testing"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/signer"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
)

func TestOnionSkins_GetBalance(t *testing.T) {
	var e error
	n3 := Gen3Nonces()
	id, confID := nonce.NewID(), nonce.NewID()
	_, ks, _ := signer.New()
	var prvs [3]*prv.Key
	for i := range prvs {
		prvs[i] = ks.Next()
	}
	var pubs [3]*pub.Key
	for i := range pubs {
		pubs[i] = pub.Derive(prvs[i])
	}
	on := Skins{}.
		GetBalance(id, confID, prvs, pubs, n3).
		Assemble()
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
	var ex *GetBalance
	var ok bool
	if ex, ok = onc.(*GetBalance); !ok {
		t.Error("did not unwrap expected type")
		t.FailNow()
	}
	if ex.ID != id {
		t.Error("ID did not decode correctly")
		t.FailNow()
	}
	for i := range ex.Ciphers {
		if ex.Ciphers[i] != on.(*GetBalance).Ciphers[i] {
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
}
