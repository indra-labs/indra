package schnorr

import (
	"bytes"
	"crypto/rand"
	"testing"
)

func TestPrivkey_ECDH(t *testing.T) {
	var err error
	var prv1, prv2 *Privkey
	var pub1, pub2 *Pubkey
	if prv1, err = GeneratePrivkey(); log.I.Chk(err) {
		t.Error(err)
	}
	pub1 = prv1.Pubkey()
	if prv2, err = GeneratePrivkey(); log.I.Chk(err) {
		t.Error(err)
	}
	pub2 = prv2.Pubkey()
	sec1 := prv1.ECDH(pub2)
	sec2 := prv2.ECDH(pub1)
	if bytes.Compare(sec1, sec2) != 0 {
		t.Error("ECDH function failed")
	}
}

func TestPrivkey_SignVerify(t *testing.T) {
	const msgSize = 4096
	message := make([]byte, msgSize)
	var err error
	var n int
	if n, err = rand.Read(message); log.E.Chk(err) && n != msgSize {
		t.Error(err)
	}
	var prv1 *Privkey
	if prv1, err = GeneratePrivkey(); log.I.Chk(err) {
		t.Error(err)
	}
	var sig *Signature
	if sig, err = prv1.Sign(message); log.I.Chk(err) {
		t.Error(err)
	}
	var pub1 *Pubkey
	pub1 = prv1.Pubkey()
	if !sig.Verify(message, pub1) {
		t.Error(err)
	}
}

func TestPrivkey_Serialize(t *testing.T) {
	var err error
	var prv1, prv2 *Privkey
	if prv1, err = GeneratePrivkey(); log.I.Chk(err) {
		t.Error(err)
	}
	pb1 := prv1.Serialize()
	prv2 = PrivkeyFromBytes(pb1[:])
	pb2 := prv2.Serialize()
	if bytes.Compare(pb1[:], pb2[:]) != 0 {
		t.Error("Privkey serialise failed")
	}
}

func TestPubkey_Serialize(t *testing.T) {
	var err error
	var prv1 *Privkey
	var pub1 *Pubkey
	if prv1, err = GeneratePrivkey(); log.I.Chk(err) {
		t.Error(err)
	}
	pb1 := prv1.Pubkey().Serialize()
	pub1, err = PubkeyFromBytes(pb1[:])
	pb2 := pub1.Serialize()
	if bytes.Compare(pb1[:], pb2[:]) != 0 {
		t.Error("Pubkey serialise failed")
	}

}

func TestSignature_Serialize(t *testing.T) {
	const msgSize = 4096
	message := make([]byte, msgSize)
	var err error
	var n int
	if n, err = rand.Read(message); log.E.Chk(err) && n != msgSize {
		t.Error(err)
	}
	var prv1 *Privkey
	if prv1, err = GeneratePrivkey(); log.I.Chk(err) {
		t.Error(err)
	}
	var sig1, sig2 *Signature
	if sig1, err = prv1.Sign(message); log.I.Chk(err) {
		t.Error(err)
	}
	sig1B := sig1.Serialize()
	if sig2, err = ParseSignature(sig1B[:]); log.I.Chk(err) {
		t.Error(err)
	}
	sig2B := sig2.Serialize()
	if bytes.Compare(sig1B[:], sig2B[:]) != 0 {
		t.Error(err)
	}

}
