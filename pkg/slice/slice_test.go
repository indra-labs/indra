package slice

import (
	"bytes"
	"crypto/rand"
	"errors"
	"testing"

	"github.com/Indra-Labs/indra/pkg/blake3"
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
	segSize := 1472
	var msg []byte
	var hash blake3.Hash
	var e error
	if msg, hash, e = GenerateTestMessage(msgSize); check(e) {
		t.Error(e)
	}
	segs := Segment(msg, segSize)
	pkt := Cat(segs...)[:len(msg)]
	hash2 := blake3.Single(pkt)
	if bytes.Compare(hash, hash2) != 0 {
		t.Error(errors.New("message did not decode" +
			" correctly"))
	}

}

func GenerateTestMessage(msgSize int) (msg []byte, hash blake3.Hash, e error) {
	msg = make([]byte, msgSize)
	var n int
	if n, e = rand.Read(msg); check(e) && n != msgSize {
		return
	}
	copy(msg, "payload")
	hash = blake3.Single(msg)
	return
}
