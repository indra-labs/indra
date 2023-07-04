package ad

import (
	"github.com/indra-labs/indra"
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/engine/coding"
	"github.com/indra-labs/indra/pkg/onions/reg"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/splice"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	if indra.CI == "false" {
		log2.SetLogLevel(log2.Trace)
	}
	var e error
	pr, _, _ := crypto.NewSigner()
	id := nonce.NewID()
	aa := New(id, pr, time.Now().Add(time.Hour))
	s := splice.New(aa.Len())
	if e = aa.Encode(s); fails(e) {
		t.FailNow()
	}
	s.SetCursor(0)
	var onc coding.Codec
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
	if !ad.Validate() {
		t.Errorf("received ad did not validate")
		t.FailNow()
	}
}
