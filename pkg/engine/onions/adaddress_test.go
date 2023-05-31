package onions

import (
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/engine/coding"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/splice"
	"github.com/multiformats/go-multiaddr"
	"testing"
)

func TestAdAddress(t *testing.T) {
	log2.SetLogLevel(log2.Trace)
	var e error
	pr, _, _ := crypto.NewSigner()
	id := nonce.NewID()
	var ma multiaddr.Multiaddr
	if ma, e = multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/4242"); fails(e) {
		t.FailNow()
	}
	aa := NewAddressAd(id, pr, ma)
	log.D.S("ad", aa)
	s := splice.New(aa.Len())
	if e = aa.Encode(s); fails(e) {
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
	var ad *AddressAd
	var ok bool
	if ad, ok = onc.(*AddressAd); !ok {
		t.Error("did not unwrap expected type")
		t.FailNow()
	}
	if ad.ID != aa.ID {
		t.Errorf("ID did not decode correctly")
		t.FailNow()
	}
	if ad.Addr.String() != aa.Addr.String() {
		t.Errorf("address did not decode correctly")
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
