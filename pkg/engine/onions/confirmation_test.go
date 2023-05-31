package onions

import (
	"testing"

	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/engine/coding"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
)

func TestOnions_Confirmation(t *testing.T) {
	log2.SetLogLevel(log2.Debug)
	id := nonce.NewID()
	var load byte = 128
	on := Skins{}.
		Confirmation(id, load).
		Assemble()
	s := Encode(on)
	s.SetCursor(0)
	var onc coding.Codec
	if onc = Recognise(s); onc == nil {
		
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
	if ci.Load != load {
		t.Error("load did not decode correctly")
		t.FailNow()
	}
}
