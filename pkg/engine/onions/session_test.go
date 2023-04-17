package onions

import (
	"testing"
	
	"git-indra.lan/indra-labs/indra/pkg/engine/coding"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
)

func TestOnionSkins_Session(t *testing.T) {
	log2.SetLogLevel(log2.Trace)
	sess := NewSessionKeys(1)
	on := Skins{}.
		Session(sess).
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
	var ci *Session
	var ok bool
	if ci, ok = onc.(*Session); !ok {
		t.Error("did not unwrap expected type")
		t.FailNow()
	}
	if !ci.Header.Key.Equals(&sess.Header.Key) {
		t.Error("header key did not unwrap correctly")
		t.FailNow()
	}
	if !ci.Payload.Key.Equals(&sess.Payload.Key) {
		t.Error("payload key did not unwrap correctly")
		t.FailNow()
	}
}
