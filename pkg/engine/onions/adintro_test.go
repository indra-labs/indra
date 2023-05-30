package onions

import (
	"testing"
	"time"

	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/engine/coding"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/slice"
)

func TestOnionSkins_Intro(t *testing.T) {
	log2.SetLogLevel(log2.Debug)
	var e error
	pr, ks, _ := crypto.NewSigner()
	id := nonce.NewID()
	in := NewIntro(id, pr, slice.GenerateRandomAddrPortIPv6(), 0, 0, time.Now().Add(time.Hour))
	var prvs crypto.Privs
	for i := range prvs {
		prvs[i] = ks.Next()
	}
	var pubs crypto.Pubs
	for i := range pubs {
		pubs[i] = crypto.DerivePub(prvs[i])
	}
	on1 := Skins{}.
		Intro(id, pr, in.AddrPort, time.Now().Add(time.Hour))
	on1 = append(on1, &End{})
	on := on1.Assemble()
	s := Encode(on)
	log.D.S(s.GetAll().ToBytes())
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
