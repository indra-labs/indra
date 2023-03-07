package engine

import (
	"reflect"
	"testing"
	
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/util/octet"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

func TestOnionSkins_Session(t *testing.T) {
	log2.SetLogLevel(log2.Trace)
	sess := NewSessionKeys(1)
	on := Skins{}.
		Session(sess).
		Assemble()
	onb := Encode(on)
	c := slice.NewCursor()
	var onc Onion
	s := octet.Load(onb, c)
	if onc = Recognise(s); onc == nil {
		t.Error("did not unwrap")
		t.FailNow()
	}
	log.I.Ln(reflect.TypeOf(onc))
	if e := onc.Decode(s); check(e) {
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
