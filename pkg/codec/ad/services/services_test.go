package services

import (
	"github.com/indra-labs/indra"
	"github.com/indra-labs/indra/pkg/codec"
	"github.com/indra-labs/indra/pkg/codec/reg"
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/splice"
	"testing"
	"time"
)

func TestServiceAd(t *testing.T) {
	if indra.CI == "false" {
		log2.SetLogLevel(log2.Trace)
	}
	var e error
	pr, _, _ := crypto.NewSigner()
	id := nonce.NewID()
	sv := New(id, pr, []Service{{80, 62346}, {443, 42216}}, time.Now().Add(time.Hour))
	log.D.S("service", sv)
	s := splice.New(sv.Len())
	if e = sv.Encode(s); fails(e) {
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
	log.D.S(onc)
	var svcAd *Ad
	var ok bool
	if svcAd, ok = onc.(*Ad); !ok {
		t.Error("did not unwrap expected type")
		t.FailNow()
	}
	if len(sv.Services) != len(svcAd.Services) {
		t.Errorf("number of services incorrectly decoded")
		t.FailNow()
	}
	for i := range sv.Services {
		if svcAd.Services[i].RelayRate != sv.Services[i].RelayRate {
			t.Errorf("relay rate did not decode correctly")
			t.FailNow()
		}
		if svcAd.Services[i].Port != sv.Services[i].Port {
			t.Errorf("port did not decode correctly")
			t.FailNow()
		}
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
