package sig

import (
	"crypto/rand"
	"errors"
	mrand "math/rand"
	"testing"

	"github.com/indra-labs/indra/pkg/key/prv"
	"github.com/indra-labs/indra/pkg/key/pub"
	"github.com/indra-labs/indra/pkg/sha256"
)

func TestSignRecover(t *testing.T) {
	msgSize := mrand.Intn(3072) + 1024
	payload := make([]byte, msgSize)
	var e error
	var n int
	if n, e = rand.Read(payload); log.E.Chk(e) && n != msgSize {
		t.Error(e)
	}
	var prv1 *prv.Key
	var pub1, rec1 *pub.Key
	if prv1, e = prv.GenerateKey(); check(e) {
		t.Error(e)
	}
	pub1 = pub.Derive(prv1)
	var s Bytes
	hash := sha256.Single(payload)
	if s, e = Sign(prv1, hash); check(e) {
		t.Error(e)
	}
	if rec1, e = s.Recover(hash); check(e) {
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
	var prv1 *prv.Key
	var pub1, rec1 *pub.Key
	if prv1, e = prv.GenerateKey(); check(e) {
		t.Error(e)
	}
	pub1 = pub.Derive(prv1)
	var s Bytes
	hash := sha256.Single(payload)
	if s, e = Sign(prv1, hash); check(e) {
		t.Error(e)
	}
	copy(payload, make([]byte, 10))
	hash2 := sha256.Single(payload)
	if rec1, e = s.Recover(hash2); check(e) {
		t.Error(e)
	}
	if pub1.Equals(rec1) {
		t.Error(errors.New("recovery extracted the same key"))
	}
}
