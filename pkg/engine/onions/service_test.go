package onions

import (
	"testing"
	"time"

	"git-indra.lan/indra-labs/indra/pkg/crypto"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/engine/coding"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
)

func TestOnionSkins_Service(t *testing.T) {
	log2.SetLogLevel(log2.Debug)
	var e error
	pr, ks, _ := crypto.NewSigner()
	id := nonce.NewID()
	// in := NewAddr(id, pr, time.Now().Add(time.Hour))
	var prvs crypto.Privs
	for i := range prvs {
		prvs[i] = ks.Next()
	}
	var pubs crypto.Pubs
	for i := range pubs {
		pubs[i] = crypto.DerivePub(prvs[i])
	}
	on1 := Skins{}.
		Addr(id, pr, time.Now().Add(time.Hour))
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
	var peer *Addr
	var ok bool
	if peer, ok = onc.(*Addr); !ok {
		t.Error("did not unwrap expected type")
		t.FailNow()
	}
	if !peer.Validate() {
		t.Errorf("received Addr did not validate")
		t.FailNow()
	}
}
