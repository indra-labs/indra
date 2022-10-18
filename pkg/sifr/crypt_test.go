package sifr

import (
	"bytes"
	"crypto/rand"
	"testing"

	"github.com/Indra-Labs/indra/pkg/schnorr"
	"github.com/Indra-Labs/indra/pkg/sha256"
)

func TestEncryptDecryptMessage(t *testing.T) {
	const msgSize = 4096
	payload := make([]byte, msgSize)
	var e error
	var n int
	if n, e = rand.Read(payload); log.E.Chk(e) && n != msgSize {
		t.Error(e)
	}
	messageHash := sha256.Hash(payload)
	var prv1, prv2 *schnorr.Privkey
	var pub1, pub2 *schnorr.Pubkey
	if prv1, e = schnorr.GeneratePrivkey(); log.I.Chk(e) {
		t.Error(e)
	}
	pub1 = prv1.Pubkey()
	var message *Message
	if message, e = NewMessage(payload, prv1); log.I.Chk(e) {
		t.Error(e)
	}
	log.I.S(message.Signature)
	if prv2, e = schnorr.GeneratePrivkey(); log.I.Chk(e) {
		t.Error(e)
	}
	pub2 = prv2.Pubkey()
	secret1 := prv1.ECDH(pub2)
	secret2 := prv2.ECDH(pub1)
	var em *Crypt
	em, e = NewCrypt(message, secret1)
	if log.E.Chk(e) {
		t.Error(e)
	}
	var decryptMessage *Message
	if decryptMessage, e = DecryptMessage(secret2,
		em.Serialize()); log.E.Chk(e) {
		t.Error(e)
	}
	log.I.S(decryptMessage.Signature)
	if e = decryptMessage.Verify(pub1); log.E.Chk(e) {
		t.Error("message failed verification with pubkey")
	}
	decryptHash := sha256.Hash(decryptMessage.Payload)
	if bytes.Compare(messageHash, decryptHash) != 0 {
		t.Error("encryption/decryption failed")
	}
}
