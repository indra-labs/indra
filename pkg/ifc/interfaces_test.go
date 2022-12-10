package ifc

import (
	"testing"

	"github.com/Indra-Labs/indra/pkg/sha256"
	"github.com/Indra-Labs/indra/pkg/testutils"
)

func TestMessage_ToU64Slice(t *testing.T) {
	var e error
	var msg1 Message
	if msg1, _, e = testutils.GenerateTestMessage(33); check(e) {
		t.Error(e)
		t.FailNow()
	}
	log.I.S(msg1)
	uMsg1 := msg1.ToU64Slice()
	umsg1 := uMsg1.ToMessage()
	_ = umsg1
}

func TestU64Slice_XOR(t *testing.T) {
	var e error
	var msg1 Message
	if msg1, _, e = testutils.GenerateTestMessage(33); check(e) {
		t.Error(e)
		t.FailNow()
	}
	hash1 := sha256.Single(msg1)
	uMsg1 := msg1.ToU64Slice()
	var msg2 Message
	if msg2, _, e = testutils.GenerateTestMessage(33); check(e) {
		t.Error(e)
		t.FailNow()
	}
	// log.I.S(msg2)
	uMsg2 := msg2.ToU64Slice()
	var msg3 Message
	if msg3, _, e = testutils.GenerateTestMessage(33); check(e) {
		t.Error(e)
		t.FailNow()
	}
	// log.I.S(msg3)
	uMsg3 := msg3.ToU64Slice()
	uMsg1.XOR(uMsg2)
	uMsg1.XOR(uMsg3)
	uMsg1.XOR(uMsg2)
	uMsg1.XOR(uMsg3)
	hash2 := sha256.Single(uMsg1.ToMessage())
	if !hash1.Equals(hash2) {
		t.Error("XOR failed")
		t.FailNow()
	}
}
