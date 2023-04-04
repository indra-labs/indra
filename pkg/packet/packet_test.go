package packet

import (
	"crypto/rand"
	"errors"
	"testing"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/cloak"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/util/tests"
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
	var sp, rp *prv.Key
	var sP, rP *pub.Key
	if sp, rp, sP, rP, e = tests.GenerateTestKeyPairs(); fails(e) {
		t.FailNow()
	}
	addr := rP
	var pkt []byte
	params := PacketParams{
		To:     addr,
		From:   sp,
		Data:   payload,
		Seq:    234,
		Parity: 64,
		Length: msgSize,
	}
	if pkt, e = Encode(params); fails(e) {
		t.Error(e)
	}
	var from *pub.Key
	var to cloak.PubKey
	_ = to
	if from, to, e = GetKeys(pkt); fails(e) {
		t.Error(e)
	}
	if !sP.ToBytes().Equals(from.ToBytes()) {
		t.Error(e)
	}
	var f *Packet
	if f, e = Decode(pkt, from, rp); fails(e) {
		t.Error(e)
	}
	dHash := sha256.Single(f.Data)
	if pHash != dHash {
		t.Error(errors.New("encode/decode unsuccessful"))
	}
}
