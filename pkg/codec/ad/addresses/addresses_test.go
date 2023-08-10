package addresses

import (
	"git.indra-labs.org/dev/ind"
	"git.indra-labs.org/dev/ind/pkg/codec"
	"git.indra-labs.org/dev/ind/pkg/codec/reg"
	"git.indra-labs.org/dev/ind/pkg/crypto"
	"git.indra-labs.org/dev/ind/pkg/crypto/nonce"
	log2 "git.indra-labs.org/dev/ind/pkg/proc/log"
	"git.indra-labs.org/dev/ind/pkg/util/splice"
	"github.com/multiformats/go-multiaddr"
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
	aa := New(id, pr, []multiaddr.Multiaddr{ma4, ma6},
		time.Now().Add(time.Hour*24*7))
	l := aa.Len()
	log.I.Ln("l", l)
	s := splice.New(l)
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
	log.D.S("received", ad)
	if ad.ID != aa.ID {
		t.Errorf("ID did not decode correctly")
		t.FailNow()
	}
	for i := range ad.Addresses {
		if ad.Addresses[i].String() != aa.Addresses[i].String() {
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
