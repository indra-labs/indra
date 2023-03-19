package ngin

import (
	"testing"
	"time"
	
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

func TestOnionSkins_Delay(t *testing.T) {
	log2.SetLogLevel(log2.Trace)
	dur := time.Second
	on := Skins{}.
		Delay(dur).
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
	var dl *Delay
	var ok bool
	if dl, ok = onc.(*Delay); !ok {
		t.Error("did not decode expected type")
		t.FailNow()
	}
	if dl.Duration != dur {
		t.Error("did not unwrap expected duration")
		t.FailNow()
	}
}
