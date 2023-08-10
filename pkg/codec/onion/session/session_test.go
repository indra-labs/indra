package session

import (
	"git.indra-labs.org/dev/ind"
	"git.indra-labs.org/dev/ind/pkg/codec"
	"git.indra-labs.org/dev/ind/pkg/codec/reg"
	"testing"

	log2 "git.indra-labs.org/dev/ind/pkg/proc/log"
)

func TestOnions_Session(t *testing.T) {
	if indra.CI == "false" {
		log2.SetLogLevel(log2.Debug)
	}
	sess := New(1)
	ss := sess.(*Session)
	s := codec.Encode(sess)
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
	var ci *Session
	var ok bool
	if ci, ok = onc.(*Session); !ok {
		t.Error("did not unwrap expected type")
		t.FailNow()
	}
	if !ci.Header.Prv.Equals(ss.Header.Prv) {
		t.Error("header key did not unwrap correctly")
		t.FailNow()
	}
	if !ci.Payload.Prv.Equals(ss.Payload.Prv) {
		t.Error("payload key did not unwrap correctly")
		t.FailNow()
	}
}
