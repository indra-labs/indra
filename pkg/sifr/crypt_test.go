package sifr

import (
	"bytes"
	"crypto/rand"
	mrand "math/rand"
	"runtime"
	"sync"
	"testing"

	"github.com/Indra-Labs/indra/pkg/mesg"
	"github.com/Indra-Labs/indra/pkg/schnorr"
	"github.com/Indra-Labs/indra/pkg/sha256"
)

func TestEncryptDecryptMessage(t *testing.T) {
	// Use a random length every time to make sure padding works.
	msgSize := mrand.Intn(3072) + 1024
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
	var message *mesg.Message
	if message, e = mesg.New(payload, prv1); log.I.Chk(e) {
		t.Error(e)
	}
	if e = message.Verify(pub1); log.E.Chk(e) {
		t.Error("message failed verification with pubkey")
	}
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
	serial := em.Serialize()
	var cr *Crypt
	if cr, e = DeserializeCrypt(serial); log.E.Chk(e) {
		t.Error(e)
	}
	var decryptMessage *mesg.Message
	if decryptMessage, e = cr.Decrypt(secret2); log.E.Chk(e) {
		t.Error(e)
	}
	if e = decryptMessage.Verify(pub1); log.E.Chk(e) {
		t.Error("message failed verification with pubkey")
	}
	decryptHash := sha256.Hash(decryptMessage.Payload)
	if bytes.Compare(messageHash, decryptHash) != 0 {
		t.Error("encryption/decryption failed")
	}
}

// This benchmark runs one thread per core to count total throughput.
//
// The following results on the author's machine work out to approximately 40
// Gb/s:
//
// goos: linux
// goarch: amd64
// pkg: github.com/Indra-Labs/indra/pkg/sifr
// cpu: AMD Ryzen 7 5800H with Radeon Graphics
// BenchmarkEncryptDecryptMessage
// BenchmarkEncryptDecryptMessage-16    	    1245	    868467 ns/op
//
// Which should be a lot more than can be actually done on a real network. This
// includes both key generation steps and verifying the decryption worked.
func BenchmarkEncryptDecryptMessage(b *testing.B) {
	// Use a random length every time to make sure padding works.
	msgSize := mrand.Intn(3072) + 1024
	payload := make([]byte, msgSize)
	var e error
	var n int
	if n, e = rand.Read(payload); log.E.Chk(e) && n != msgSize {
	}
	ncpu := runtime.NumCPU()
	var wg sync.WaitGroup
	for n := 0; n < b.N; n++ {
		for i := 0; i < ncpu; i++ {
			go func() {
				wg.Add(1)
				var prv1, prv2 *schnorr.Privkey
				var pub1, pub2 *schnorr.Pubkey
				if prv1, e = schnorr.GeneratePrivkey(); log.I.Chk(e) {
				}
				pub1 = prv1.Pubkey()
				var message *mesg.Message
				if message, e = mesg.New(payload, prv1); log.I.Chk(e) {
				}
				if prv2, e = schnorr.GeneratePrivkey(); log.I.Chk(e) {
				}
				pub2 = prv2.Pubkey()
				secret1 := prv1.ECDH(pub2)
				var em *Crypt
				em, e = NewCrypt(message, secret1)
				if log.E.Chk(e) {
				}
				serial := em.Serialize()
				var cr *Crypt
				if cr, e = DeserializeCrypt(serial); log.E.Chk(e) {
				}
				var dc *mesg.Message
				if dc, e = cr.Decrypt(secret1); log.E.Chk(e) {
				}
				if e = dc.Verify(pub1); log.E.Chk(e) {
				}
				wg.Done()
			}()
		}
	}
	wg.Wait()
}
