package crypt

import (
	"github.com/indra-labs/indra"
	"github.com/indra-labs/indra/pkg/codec"
	"github.com/indra-labs/indra/pkg/codec/onion/confirmation"
	"github.com/indra-labs/indra/pkg/codec/ont"
	"github.com/indra-labs/indra/pkg/codec/reg"
	"testing"

	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
)

func TestOnions_SimpleCrypt(t *testing.T) {
	if indra.CI == "false" {
		log2.SetLogLevel(log2.Debug)
	}
	var e error
	n := nonce.NewID()
	n1 := nonce.New()
	prv1, prv2 := crypto.GetTwoPrvKeys()
	pub1, pub2 := crypto.DerivePub(prv1), crypto.DerivePub(prv2)
	on := ont.Assemble([]ont.Onion{
		New(pub1, pub2, prv2, n1, 0),
		confirmation.New(n),
	})
	s := ont.Encode(on)
	s.SetCursor(0)
	log.D.S("encoded, encrypted onion:\n", s.GetAll().ToBytes())
	var oncr codec.Codec
	if oncr = reg.Recognise(s); oncr == nil {
		t.Error("did not unwrap")
		t.FailNow()
	}
	if e = oncr.Decode(s); fails(e) {
		t.Error("did not decode")
		t.FailNow()
	}
	oncr.(*Crypt).Decrypt(prv1, s)
	var oncn codec.Codec
	if oncn = reg.Recognise(s); oncn == nil {
		t.Error("did not unwrap")
		t.FailNow()
	}
	if e = oncn.Decode(s); fails(e) {
		t.Error("did not decode")
		t.FailNow()
	}
	if cn, ok := oncn.(*confirmation.Confirmation); !ok {
		t.Error("did not get expected confirmation")
		t.FailNow()
	} else {
		if cn.ID != n {
			t.Error("did not get expected confirmation Keys")
			t.FailNow()
		}
	}
}
