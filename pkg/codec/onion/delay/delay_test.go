package delay

import (
	"git.indra-labs.org/dev/ind/pkg/codec"
	"git.indra-labs.org/dev/ind/pkg/codec/ont"
	"git.indra-labs.org/dev/ind/pkg/codec/reg"
	"git.indra-labs.org/dev/ind/pkg/util/ci"
	"testing"
	"time"
)

func TestOnions_Delay(t *testing.T) {
	ci.TraceIfNot()
	dur := time.Second
	on := ont.Assemble([]ont.Onion{New(dur)})
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
