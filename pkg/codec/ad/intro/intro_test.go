package intro

import (
	"git.indra-labs.org/dev/ind/pkg/codec"
	"git.indra-labs.org/dev/ind/pkg/codec/reg"
	"git.indra-labs.org/dev/ind/pkg/util/ci"
	"git.indra-labs.org/dev/ind/pkg/util/splice"
	"testing"
	"time"
	
	"git.indra-labs.org/dev/ind/pkg/crypto"
	"git.indra-labs.org/dev/ind/pkg/crypto/nonce"
)

func TestNew(t *testing.T) {
	ci.TraceIfNot()
	var e error
	pr, ks, _ := crypto.NewSigner()
	introducer := ks.Next()
	id := nonce.NewID()
	in := New(id, pr, crypto.DerivePub(introducer),
		20000, 80, time.Now().Add(time.Hour))
	s := splice.New(in.Len())
	if e = in.Encode(s); fails(e) {
		t.FailNow()
	}
	s.SetCursor(0)
	var onc codec.Codec
	if onc = reg.Recognise(s); onc == nil {
		t.Error("did not unwrap")
		t.FailNow()
	}
	if e = onc.Decode(s); fails(e) {
		t.Error("did not decode")
		t.FailNow()
	}
	var ad *Ad
	var ok bool
	if ad, ok = onc.(*Ad); !ok {
		t.Error("did not unwrap expected type")
		t.FailNow()
	}
	log.D.S(ad)
	if ad.ID != in.ID {
		t.Errorf("ID did not decode correctly")
		t.FailNow()
	}
	if ad.Port != in.Port {
		t.Errorf("port did not decode correctly")
		t.FailNow()
	}
	if ad.RelayRate != in.RelayRate {
		t.Errorf("relay rate did not decode correctly")
		t.FailNow()
	}
	if ad.Expiry.Unix() != in.Expiry.Unix() {
		t.Errorf("expiry did not decode correctly")
		t.FailNow()
	}
	if !ad.Key.Equals(crypto.DerivePub(pr)) {
		t.Errorf("public key did not decode correctly")
		t.FailNow()
	}
	if !ad.Introducer.Equals(crypto.DerivePub(introducer)) {
		t.Errorf("public key did not decode correctly")
		t.FailNow()
	}
	if !ad.Validate() {
		t.Errorf("received intro did not validate")
		t.FailNow()
	}
}
