package onions

import (
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/engine/coding"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/splice"
	"testing"
)

func TestServiceAd(t *testing.T) {
	log2.SetLogLevel(log2.Trace)
	var e error
	pr, _, _ := crypto.NewSigner()
	id := nonce.NewID()
	sv := NewServiceAd(id, pr, 20000, 80)
	log.D.S("service", sv)
	s := splice.New(sv.Len())
	if e = sv.Encode(s); fails(e) {
		t.FailNow()
	}
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
	log.D.S(onc)
	var svcAd *ServiceAd
	var ok bool
	if svcAd, ok = onc.(*ServiceAd); !ok {
		t.Error("did not unwrap expected type")
		t.FailNow()
	}
	if svcAd.RelayRate != sv.RelayRate {
		t.Errorf("relay rate did not decode correctly")
		t.FailNow()
	}
	if svcAd.Port != sv.Port {
		t.Errorf("port did not decode correctly")
		t.FailNow()
	}
	if !svcAd.Key.Equals(crypto.DerivePub(pr)) {
		t.Errorf("public key did not decode correctly")
		t.FailNow()
	}
	if !svcAd.Validate() {
		t.Errorf("received service ad did not validate")
		t.FailNow()
	}
}
