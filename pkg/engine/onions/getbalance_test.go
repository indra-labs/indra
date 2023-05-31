package onions

import (
	"testing"

	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/engine/coding"
	"github.com/indra-labs/indra/pkg/engine/sessions"
)

func TestOnions_GetBalance(t *testing.T) {
	var e error
	n3 := crypto.Gen3Nonces()
	id := nonce.NewID()
	_, ks, _ := crypto.NewSigner()
	var prvs crypto.Privs
	for i := range prvs {
		prvs[i] = ks.Next()
	}
	var pubs crypto.Pubs
	for i := range pubs {
		pubs[i] = crypto.DerivePub(prvs[i])
	}
	ep := &ExitPoint{
		Routing: &Routing{
			Sessions: [3]*sessions.Data{},
			Keys:     prvs,
			Nonces:   n3,
		},
		ReturnPubs: pubs,
	}
	on := Skins{}.
		GetBalance(id, ep).
		Assemble()
	s := Encode(on)
	s.SetCursor(0)
	var onc coding.Codec
	if onc = Recognise(s); onc == nil {
		t.Error("did not unwrap")
		t.FailNow()
	}
	if e = onc.Decode(s); fails(e) {
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
		t.Error("Keys did not decode correctly")
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
