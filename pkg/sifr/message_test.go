package sifr

import (
	"bytes"
	"crypto/rand"
	"testing"

	"github.com/Indra-Labs/indra/pkg/schnorr"
)

func TestEncryptDecryptMessage(t *testing.T) {
	const msgSize = 4096
	message := make([]byte, msgSize)
	var e error
	var n int
	if n, e = rand.Read(message); log.E.Chk(e) && n != msgSize {
		t.Error(e)
	}
	messageHash := schnorr.SHA256(message)
	var prv1, prv2 *schnorr.Privkey
	var pub1, pub2 *schnorr.Pubkey
	if prv1, e = schnorr.GeneratePrivkey(); log.I.Chk(e) {
		t.Error(e)
	}
	pub1 = prv1.Pubkey()
	if prv2, e = schnorr.GeneratePrivkey(); log.I.Chk(e) {
		t.Error(e)
	}
	pub2 = prv2.Pubkey()
	secret := prv1.ECDH(pub2)
	var em *EncryptedMessage
	em, e = EncryptMessage(secret, message, prv1)
	if log.E.Chk(e) {
		t.Error(e)
	}
	var decryptMessage []byte
	decryptMessage, e = DecryptMessage(secret, em.Serialize(), pub1)
	if log.E.Chk(e) {
		t.Error(e)
	}
	decryptHash := schnorr.SHA256(decryptMessage)
	if bytes.Compare(messageHash, decryptHash) != 0 {
		t.Error("encryption/decryption failed")
	}
}
