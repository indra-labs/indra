package packet

import (
	"crypto/rand"
	"errors"
	"testing"

	"git.indra-labs.org/dev/ind/pkg/crypto"
	"git.indra-labs.org/dev/ind/pkg/crypto/nonce"
	"git.indra-labs.org/dev/ind/pkg/crypto/sha256"
)

func TestEncode_Decode(t *testing.T) {
	msgSize := 1382
	payload := make([]byte, msgSize)
	var e error
	var n int
	if n, e = rand.Read(payload); fails(e) && n != msgSize {
		t.Error(e)
	}
	payload = append([]byte("payload"), payload...)
	pHash := sha256.Single(payload)
	var sp, rp *crypto.Prv
	var sP, rP *crypto.Pub
	if sp, rp, sP, rP, e = crypto.GenerateTestKeyPairs(); fails(e) {
		t.FailNow()
	}
	addr := rP
	var pkt []byte
	params := &PacketParams{
		To:     addr,
		From:   sp,
		Data:   payload,
		Seq:    234,
		Parity: 64,
		Length: msgSize,
	}
	if pkt, e = EncodePacket(params); fails(e) {
		t.Error(e)
	}
	var from *crypto.Pub
	var to crypto.CloakedPubKey
	_ = to
	var iv nonce.IV
	if from, to, iv, e = GetKeysFromPacket(pkt); fails(e) {
		t.Error(e)
	}
	if !sP.ToBytes().Equals(from.ToBytes()) {
		t.Error(e)
	}
	var f *Packet
	if f, e = DecodePacket(pkt, from, rp, iv); fails(e) {
		t.Error(e)
	}
	dHash := sha256.Single(f.Data)
	if pHash != dHash {
		t.Error(errors.New("encode/decode unsuccessful"))
	}
}
