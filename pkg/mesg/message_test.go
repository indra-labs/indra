package mesg

import (
	"bytes"
	"crypto/rand"
	"errors"
	mrand "math/rand"
	"testing"

	"github.com/Indra-Labs/indra/pkg/keys"
)

func TestSerializeDeserializeMessage(t *testing.T) {
	// Use a random length every time to ensure padding works.
	msgSize := mrand.Intn(3072) + 1024
	payload := make([]byte, msgSize)
	var e error
	var n int
	if n, e = rand.Read(payload); log.E.Chk(e) && n != msgSize {
		t.Error(e)
	}
	var prv1 *keys.Privkey
	var pub1 *keys.Pubkey
	if prv1, e = keys.GeneratePrivkey(); log.I.Chk(e) {
		t.Error(e)
	}
	pub1 = prv1.Pubkey()
	var msg *Message
	if msg, e = New(payload, prv1); log.E.Chk(e) {
		t.Error(e)
	}
	if e = msg.Verify(pub1); log.E.Chk(e) {
		t.Error(e)
	}
	serial := msg.Serialize()
	var msg2 *Message
	if msg2, e = DeserializeMessage(serial); log.E.Chk(e) {
		t.Error(e)
	}
	if e = msg2.Verify(pub1); log.E.Chk(e) {
		t.Error(e)
	}
	if bytes.Compare(msg.Payload, msg2.Payload) != 0 {
		log.I.S(msg.Payload, msg2.Payload)
		t.Error(errors.New("failed to deserialize"))
	}
}

func BenchmarkSerializeDeserializeMessage(b *testing.B) {
	msgSize := mrand.Intn(3072) + 1024
	payload := make([]byte, msgSize)
	var e error
	var n int
	if n, e = rand.Read(payload); log.E.Chk(e) && n != msgSize {

	}
	var prv1 *keys.Privkey
	if prv1, e = keys.GeneratePrivkey(); log.I.Chk(e) {

	}
	// Use a random length every time to ensure padding works.
	var msg *Message
	if msg, e = New(payload, prv1); log.E.Chk(e) {
	}
	for n := 0; n < b.N; n++ {
		serial := msg.Serialize()
		if _, e = DeserializeMessage(serial); log.E.Chk(e) {
			continue
		}
	}
}
