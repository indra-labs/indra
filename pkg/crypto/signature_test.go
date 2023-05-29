package crypto

import (
	"crypto/rand"
	"errors"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	mrand "math/rand"
	"testing"
)

func TestSignRecover(t *testing.T) {
	msgSize := mrand.Intn(3072) + 1024
	payload := make([]byte, msgSize)
	var e error
	var n int
	if n, e = rand.Read(payload); log.E.Chk(e) && n != msgSize {
		t.Error(e)
	}
	var prv1 *Prv
	var pub1, rec1 *Pub
	if prv1, e = GeneratePrvKey(); fails(e) {
		t.Error(e)
	}
	pub1 = DerivePub(prv1)
	var s SigBytes
	hash := sha256.Single(payload)
	if s, e = Sign(prv1, hash); fails(e) {
		t.Error(e)
	}
	if rec1, e = s.Recover(hash); fails(e) {
		t.Error(e)
	}
	if !pub1.Equals(rec1) {
		t.Error(errors.New("recovery did not extract same key"))
	}
}

func TestSignRecoverFail(t *testing.T) {
	msgSize := mrand.Intn(3072) + 1024
	payload := make([]byte, msgSize)
	var e error
	var n int
	if n, e = rand.Read(payload); log.E.Chk(e) && n != msgSize {
		t.Error(e)
	}
	var prv1 *Prv
	var pub1, rec1 *Pub
	if prv1, e = GeneratePrvKey(); fails(e) {
		t.Error(e)
	}
	pub1 = DerivePub(prv1)
	var s SigBytes
	hash := sha256.Single(payload)
	if s, e = Sign(prv1, hash); fails(e) {
		t.Error(e)
	}
	copy(payload, make([]byte, 10))
	hash2 := sha256.Single(payload)
	if rec1, e = s.Recover(hash2); fails(e) {
		t.Error(e)
	}
	if pub1.Equals(rec1) {
		t.Error(errors.New("recovery extracted the same key"))
	}
}
