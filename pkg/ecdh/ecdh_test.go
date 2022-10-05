package ecdh

import (
	"bytes"
	"testing"
)

func TestComputeSecret(t *testing.T) {
	var err error
	var kp1, kp2 Keypair
	if kp1, err = GenerateKeypair(); log.E.Chk(err) {
		t.Error(err)
	}
	if kp2, err = GenerateKeypair(); log.E.Chk(err) {
		t.Error(err)
	}
	var sec1, sec2 []byte
	if sec1, err = kp1.ComputeSecret(kp2.Pubkey); log.E.Chk(err) {
		t.Error(err)
	}
	if sec2, err = kp2.ComputeSecret(kp1.Pubkey); log.E.Chk(err) {
		t.Error(err)
	}
	if bytes.Compare(sec1, sec2) != 0 {
		t.Error("secrets do not match")
	}
}
