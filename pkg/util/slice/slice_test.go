package slice

import (
	"errors"
	"testing"

	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/util/tests"
)

func TestSize24(t *testing.T) {
	n := 1<<24 - 1
	u := NewUint24()
	EncodeUint24(u, n)
	u2 := DecodeUint24(u)
	if n != u2 {
		t.Error("failed to encode/decode")
	}
}

func TestSegment(t *testing.T) {
	msgSize := 2 << 17
	segSize := 1382
	var msg []byte
	var hash sha256.Hash
	var e error
	if msg, hash, e = tests.GenMessage(msgSize, ""); check(e) {
		t.Error(e)
	}
	segs := Segment(msg, segSize)
	pkt := Cat(segs...)[:len(msg)]
	hash2 := sha256.Single(pkt)
	if hash != hash2 {
		t.Error(errors.New("message did not decode" +
			" correctly"))
	}

}
