package confirmation

import (
	"git.indra-labs.org/dev/ind"
	"git.indra-labs.org/dev/ind/pkg/codec"
	"git.indra-labs.org/dev/ind/pkg/codec/ont"
	"git.indra-labs.org/dev/ind/pkg/codec/reg"
	"testing"

	"git.indra-labs.org/dev/ind/pkg/crypto/nonce"
	log2 "git.indra-labs.org/dev/ind/pkg/proc/log"
)

func TestOnions_Confirmation(t *testing.T) {
	if indra.CI == "false" {
		log2.SetLogLevel(log2.Debug)
	}
	id := nonce.NewID()
	on := ont.Assemble([]ont.Onion{New(id)})
	s := codec.Encode(on)
	s.SetCursor(0)
	var onc codec.Codec
	if onc = reg.Recognise(s); onc == nil {

		t.Error("did not unwrap")
		t.FailNow()
	}
	if e := onc.Decode(s); fails(e) {
		t.Error("did not decode")
		t.FailNow()
	}
	var ci *Confirmation
	var ok bool
	if ci, ok = onc.(*Confirmation); !ok {
		t.Error("did not unwrap expected type")
		t.FailNow()
	}
	if ci.ID != id {
		t.Error("Keys did not decode correctly")
		t.FailNow()
	}
}
