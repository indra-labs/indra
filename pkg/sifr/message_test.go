package sifr

import (
	"bytes"
	"crypto/rand"
	"errors"
	mrand "math/rand"
	"testing"

	"github.com/Indra-Labs/indra/pkg/schnorr"
)

func TestSerializeDeserializeMessage(t *testing.T) {
	// Use a random length every time to ensure padding works.
	msgSize := mrand.Intn(4096) + 1024
	payload := make([]byte, msgSize)
	var e error
	var n int
	if n, e = rand.Read(payload); log.E.Chk(e) && n != msgSize {
		t.Error(e)
	}
	var prv1 *schnorr.Privkey
	var pub1 *schnorr.Pubkey
	if prv1, e = schnorr.GeneratePrivkey(); log.I.Chk(e) {
		t.Error(e)
	}
	pub1 = prv1.Pubkey()
	var msg *Message
	if msg, e = NewMessage(payload, prv1); log.E.Chk(e) {
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
