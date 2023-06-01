package intro

import (
	"github.com/indra-labs/indra/pkg/onions/reg"
	"github.com/indra-labs/indra/pkg/util/splice"
	"testing"
	"time"

	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/engine/coding"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/slice"
)

func TestIntroAd(t *testing.T) {
	log2.SetLogLevel(log2.Trace)
	var e error
	pr, _, _ := crypto.NewSigner()
	id := nonce.NewID()
	in := NewIntroAd(id, pr, slice.GenerateRandomAddrPortIPv6(),
		20000, 80, time.Now().Add(time.Hour))
	log.D.S("intro", in)
	s := splice.New(in.Len())
	if e = in.Encode(s); fails(e) {
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
	log.D.S(onc)
	var intro *Ad
	var ok bool
	if intro, ok = onc.(*Ad); !ok {
		t.Error("did not unwrap expected type")
		t.FailNow()
	}
	if intro.ID != in.ID {
		t.Errorf("ID did not decode correctly")
		t.FailNow()
	}
	if intro.Port != in.Port {
		t.Errorf("port did not decode correctly")
		t.FailNow()
	}
	if intro.RelayRate != in.RelayRate {
		t.Errorf("relay rate did not decode correctly")
		t.FailNow()
	}
	if !intro.Expiry.Equal(in.Expiry) {
		t.Errorf("expiry did not decode correctly")
		t.FailNow()
	}
	if intro.AddrPort.String() != in.AddrPort.String() {
		t.Errorf("addrport did not decode correctly")
		t.FailNow()
	}
	if !intro.Key.Equals(crypto.DerivePub(pr)) {
		t.Errorf("public key did not decode correctly")
		t.FailNow()
	}
	if !intro.Validate() {
		t.Errorf("received intro did not validate")
		t.FailNow()
	}
}
