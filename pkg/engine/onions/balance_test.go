package onions

import (
	"testing"
	
	"github.com/lightningnetwork/lnd/lnwire"
	
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/engine/coding"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
)

func TestOnionSkins_Balance(t *testing.T) {
	log2.SetLogLevel(log2.Debug)
	id := nonce.NewID()
	sats := lnwire.MilliSatoshi(10000)
	on := Skins{}.
		Balance(id, sats).
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
	var ci *Balance
	var ok bool
	if ci, ok = onc.(*Balance); !ok {
		t.Error("did not unwrap expected type")
		t.FailNow()
	}
	if ci.ID != id {
		t.Error("Keys did not decode correctly")
		t.FailNow()
	}
	if ci.MilliSatoshi != sats {
		t.Error("amount did not decode correctly")
		t.FailNow()
	}
}
