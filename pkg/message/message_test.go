package message

import (
	"bytes"
	"crypto/rand"
	"errors"
	mrand "math/rand"
	"testing"

	"github.com/Indra-Labs/indra/pkg/key/address"
	"github.com/Indra-Labs/indra/pkg/key/prv"
	"github.com/Indra-Labs/indra/pkg/key/pub"
	"github.com/Indra-Labs/indra/pkg/sha256"
)

func TestEncode_Decode(t *testing.T) {
	msgSize := mrand.Intn(3072) + 1024
	payload := make([]byte, msgSize)
	var e error
	var n int
	if n, e = rand.Read(payload); check(e) && n != msgSize {
		t.Error(e)
	}
	payload = append([]byte("payload"), payload...)
	pHash := sha256.Single(payload)
	var sp, rp, Rp *prv.Key
	var sP, rP, RP *pub.Key
	var ret pub.Bytes
	if sp, rp, Rp, sP, rP, RP, ret, e = GenerateTestKeyPairs(); check(e) {
		t.FailNow()
	}
	_, _, _ = sP, Rp, RP
	addr := address.FromPubKey(rP)
	var pkt []byte
	params := EP{
		To:     addr,
		From:   sp,
		Data:   payload,
		Seq:    234,
		Parity: 64,
		Length: msgSize,
		Return: ret,
	}
	if pkt, e = Encode(params); check(e) {
		t.Error(e)
	}
	var to address.Cloaked
	var from *pub.Key
	if to, from, e = GetKeys(pkt); check(e) {
		t.Error(e)
	}
	_ = to
	if !sP.ToBytes().Equals(from.ToBytes()) {
		t.Error(e)
	}
	rk := address.NewReceiver(rp)
	if !rk.Match(to) {
		t.Error("cloaked key incorrect")
	}
	var f *Packet
	if f, e = Decode(pkt, from, rp); check(e) {
		t.Error(e)
	}
	dHash := sha256.Single(f.Data)
	if bytes.Compare(pHash, dHash) != 0 {
		t.Error(errors.New("encode/decode unsuccessful"))
	}

}
