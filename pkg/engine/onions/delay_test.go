package onions

import (
	"github.com/indra-labs/indra/pkg/engine/onions/reg"
	"testing"
	"time"

	"github.com/indra-labs/indra/pkg/engine/coding"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
)

func TestOnions_Delay(t *testing.T) {
	log2.SetLogLevel(log2.Debug)
	dur := time.Second
	on := Skins{}.
		Delay(dur).
		Assemble()
	s := Encode(on)
	s.SetCursor(0)
	var onc coding.Codec
	if onc = reg.Recognise(s); onc == nil {
		t.Error("did not unwrap")
		t.FailNow()
	}
	if e := onc.Decode(s); fails(e) {
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
