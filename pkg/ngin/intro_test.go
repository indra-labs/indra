package ngin

import (
	"testing"
	"time"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/signer"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

func TestOnionSkins_Intro(t *testing.T) {
	log2.SetLogLevel(log2.Trace)
	var e error
	pr, ks, _ := signer.New()
	id := nonce.NewID()
	in := NewIntro(id, pr, slice.GenerateRandomAddrPortIPv6(),
		time.Now().Add(time.Hour))
	var prvs [3]*prv.Key
	for i := range prvs {
		prvs[i] = ks.Next()
	}
	var pubs [3]*pub.Key
	for i := range pubs {
		pubs[i] = pub.Derive(prvs[i])
	}
	on1 := Skins{}.
		Intro(id, pr, in.AddrPort, time.Now().Add(time.Hour))
	on1 = append(on1, &End{})
	on := on1.Assemble()
	s := Encode(on)
	log.D.S(s.GetRange(-1, -1).ToBytes())
	s.SetCursor(0)
	var onc Onion
	if onc = Recognise(s, slice.GenerateRandomAddrPortIPv6()); onc == nil {
		t.Error("did not unwrap")
		t.FailNow()
	}
	if e = onc.Decode(s); check(e) {
		t.Error("did not decode")
		t.FailNow()
	}
	log.D.S(onc)
	var intro *Intro
	var ok bool
	if intro, ok = onc.(*Intro); !ok {
		t.Error("did not unwrap expected type")
		t.FailNow()
	}
	if intro.AddrPort.String() != in.AddrPort.String() {
		t.Errorf("addrport did not decode correctly")
		t.FailNow()
	}
	if !intro.Validate() {
		t.Errorf("received intro did not validate")
		t.FailNow()
	}
}
