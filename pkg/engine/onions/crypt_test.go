package onions

import (
	"testing"

	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/engine/coding"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
)

func TestOnionSkins_SimpleCrypt(t *testing.T) {
	log2.SetLogLevel(log2.Debug)
	var e error
	n := nonce.NewID()
	n1 := nonce.New()
	prv1, prv2 := crypto.GetTwoPrvKeys(t)
	pub1, pub2 := crypto.DerivePub(prv1), crypto.DerivePub(prv2)
	on := Skins{}.
		Crypt(pub1, pub2, prv2, n1, 0).
		Confirmation(n, 0).
		Assemble()
	s := Encode(on)
	s.SetCursor(0)
	log.D.S("encoded, encrypted onion:\n", s.GetAll().ToBytes())
	var oncr coding.Codec
	if oncr = Recognise(s); oncr == nil {
		t.Error("did not unwrap")
		t.FailNow()
	}
	if e = oncr.Decode(s); fails(e) {
		t.Error("did not decode")
		t.FailNow()
	}
	oncr.(*Crypt).Decrypt(prv1, s)
	var oncn coding.Codec
	if oncn = Recognise(s); oncn == nil {
		t.Error("did not unwrap")
		t.FailNow()
	}
	if e = oncn.Decode(s); fails(e) {
		t.Error("did not decode")
		t.FailNow()
	}
	if cn, ok := oncn.(*Confirmation); !ok {
		t.Error("did not get expected confirmation")
		t.FailNow()
	} else {
		if cn.ID != n {
			t.Error("did not get expected confirmation Keys")
			t.FailNow()
		}
	}
}
