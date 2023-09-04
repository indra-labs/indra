package peer

import (
	"git.indra-labs.org/dev/ind/pkg/codec"
	"git.indra-labs.org/dev/ind/pkg/util/ci"
	"testing"
	"time"
	
	"git.indra-labs.org/dev/ind/pkg/codec/reg"
	"git.indra-labs.org/dev/ind/pkg/crypto"
	"git.indra-labs.org/dev/ind/pkg/crypto/nonce"
	"git.indra-labs.org/dev/ind/pkg/util/splice"
)

func TestNew(t *testing.T) {
	ci.TraceIfNot()
	var e error
	pr, _, _ := crypto.NewSigner()
	id := nonce.NewID()
	aa := New(id, pr, 20000, time.Now().Add(time.Hour))
	s := splice.New(aa.Len())
	if e = aa.Encode(s); fails(e) {
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
	log.T.S(ad)
	if ad.ID != aa.ID {
		t.Errorf("ID did not decode correctly")
		t.FailNow()
	}
	if ad.Expiry.Unix() != aa.Expiry.Unix() {
		t.Errorf("expiry did not decode correctly")
		t.FailNow()
	}
	if !ad.Key.Equals(crypto.DerivePub(pr)) {
		t.Errorf("public key did not decode correctly")
		t.FailNow()
	}
	if ad.RelayRate != aa.RelayRate {
		t.Errorf("received ad did not have same relay rate")
		t.FailNow()
	}
	if !ad.Validate() {
		t.Errorf("received ad did not validate")
		t.FailNow()
	}
}
