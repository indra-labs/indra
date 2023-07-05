package addresses

import (
	"github.com/indra-labs/indra"
	"github.com/indra-labs/indra/pkg/codec"
	"github.com/indra-labs/indra/pkg/codec/reg"
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/multi"
	"github.com/indra-labs/indra/pkg/util/splice"
	"github.com/multiformats/go-multiaddr"
	"net/netip"
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
	var ma4, ma6 multiaddr.Multiaddr
	if ma4, e = multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/4242"); fails(e) {
		t.FailNow()
	}
	if ma6, e = multiaddr.NewMultiaddr("/ip6/::1/tcp/4242"); fails(e) {
		t.FailNow()
	}
	var ap4, ap6 netip.AddrPort
	if ap4, e = multi.AddrToAddrPort(ma4); fails(e) {
		t.FailNow()
	}
	if ap6, e = multi.AddrToAddrPort(ma6); fails(e) {
		t.FailNow()
	}
	aa := New(id, pr, []*netip.AddrPort{&ap4, &ap6}, time.Now().Add(time.Hour*24*7))
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
	if ad.Expiry.Unix() != aa.Expiry.Unix() {
		t.Errorf("expiry did not decode correctly")
		t.FailNow()
	}
	log.D.S(ad)
	if ad.ID != aa.ID {
		t.Errorf("ID did not decode correctly")
		t.FailNow()
	}
	for i := range ad.Addresses {
		if *ad.Addresses[i] != *aa.Addresses[i] {
			t.Errorf("address did not decode correctly")
			t.FailNow()
		}
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
