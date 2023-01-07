package packet

import (
	"crypto/rand"
	"errors"
	"testing"
	"time"

	"github.com/indra-labs/indra/pkg/key/prv"
	"github.com/indra-labs/indra/pkg/key/pub"
	"github.com/indra-labs/indra/pkg/sha256"
	"github.com/indra-labs/indra/pkg/testutils"
)

func TestEncode_Decode(t *testing.T) {
	msgSize := 1382
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
	if sp, rp, sP, rP, e = testutils.GenerateTestKeyPairs(); check(e) {
		t.FailNow()
	}
	addr := rP
	var pkt []byte
	params := EP{
		To:       addr,
		From:     sp,
		Data:     payload,
		Seq:      234,
		Parity:   64,
		Deadline: time.Now().Add(time.Minute),
		Length:   msgSize,
	}
	if pkt, e = Encode(params); check(e) {
		t.Error(e)
	}
	var from *pub.Key
	if from, e = GetKeys(pkt); check(e) {
		t.Error(e)
	}
	if !sP.ToBytes().Equals(from.ToBytes()) {
		t.Error(e)
	}
	var f *Packet
	if f, e = Decode(pkt, from, rp); check(e) {
		t.Error(e)
	}
	dHash := sha256.Single(f.Data)
	if pHash != dHash {
		t.Error(errors.New("encode/decode unsuccessful"))
	}
}
