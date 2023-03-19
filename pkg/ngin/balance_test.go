package ngin

import (
	"testing"
	
	"git-indra.lan/indra-labs/lnd/lnd/lnwire"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

func TestOnionSkins_Balance(t *testing.T) {
	log2.SetLogLevel(log2.Trace)
	id, confID := nonce.NewID(), nonce.NewID()
	sats := lnwire.MilliSatoshi(10000)
	on := Skins{}.
		Balance(id, confID, sats).
		Assemble()
	s := Encode(on)
	s.SetCursor(0)
	var onc Onion
	if onc = Recognise(s, slice.GenerateRandomAddrPortIPv6()); onc == nil {
		t.Error("did not unwrap")
		t.FailNow()
	}
	if e := onc.Decode(s); check(e) {
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
		t.Error("ID did not decode correctly")
		t.FailNow()
	}
	if ci.ConfID != confID {
		t.Error("Confirmation ID did not decode correctly")
		t.FailNow()
	}
	if ci.MilliSatoshi != sats {
		t.Error("amount did not decode correctly")
		t.FailNow()
	}
}
