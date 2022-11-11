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
	var sp, rp *prv.Key
	var sP, rP *pub.Key
	if sp, rp, sP, rP, e = GenerateTestKeyPairs(); check(e) {
		t.FailNow()
	}
	addr := address.NewAddress(rP)
	var pkt []byte
	params := EP{
		To:     addr,
		From:   sp,
		Data:   payload,
		Seq:    234,
		Parity: 64,
		Length: msgSize,
	}
	if pkt, e = Encode(params); check(e) {
		t.Error(e)
	}
	// var to address.AddressBytes
	// var from *pub.Key
	// if to, from, e = GetKeys(pkt); check(e) {
	// 	t.Error(e)
	// }
	// if e = pub.Derive(rp).ToBytes().Fingerprint().Equals(to); check(e) {
	// 	t.Error(e)
	// }
	// if e = from.ToBytes().Fingerprint().
	// 	Equals(pub.Derive(sp).ToBytes().Fingerprint()); check(e) {
	// 	t.Error(e)
	// }
	var f *Packet
	if f, e = Decode(pkt, sP, rp); check(e) {
		t.Error(e)
	}
	dHash := sha256.Single(f.Data)
	if bytes.Compare(pHash, dHash) != 0 {
		t.Error(errors.New("encode/decode unsuccessful"))
	}

}
