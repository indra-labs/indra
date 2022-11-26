package segment

import (
	"bytes"
	"errors"
	"testing"

	"github.com/Indra-Labs/indra/pkg/key/address"
	"github.com/Indra-Labs/indra/pkg/key/prv"
	"github.com/Indra-Labs/indra/pkg/key/pub"
	"github.com/Indra-Labs/indra/pkg/packet"
	"github.com/Indra-Labs/indra/pkg/sha256"
	"github.com/Indra-Labs/indra/pkg/testutils"
)

func TestSplitJoin(t *testing.T) {
	msgSize := 2 << 19
	segSize := 1472
	var e error
	var payload []byte
	var pHash sha256.Hash

	if payload, pHash, e = testutils.GenerateTestMessage(msgSize); check(e) {
		t.FailNow()
	}
	var sp, rp, Rp *prv.Key
	var sP, rP, RP *pub.Key
	if sp, rp, sP, rP, e = testutils.GenerateTestKeyPairs(); check(e) {
		t.FailNow()
	}
	_, _, _, _ = sP, Rp, RP, rp
	addr := address.FromPubKey(rP)
	params := packet.EP{
		To:     addr,
		From:   sp,
		Length: len(payload),
		Data:   payload,
		Parity: 128,
	}
	var splitted [][]byte
	if splitted, e = Split(params, segSize); check(e) {
		t.Error(e)
	}
	var pkts packet.Packets
	var keys []*pub.Key
	for i := range splitted {
		var pkt *packet.Packet
		var from *pub.Key
		if _, from, e = packet.GetKeys(splitted[i]); check(e) {
			log.I.Ln(i)
			continue
		}
		if pkt, e = packet.Decode(splitted[i], from, rp); check(e) {
			t.Error(e)
		}
		pkts = append(pkts, pkt)
		keys = append(keys, from)
	}
	prev := keys[0]
	// check all keys are the same
	for _, k := range keys[1:] {
		if !prev.Equals(k) {
			t.Error(e)
		}
		prev = k
	}
	var msg []byte
	if msg, e = Join(pkts); check(e) {
		t.Error(e)
	}
	rHash := sha256.Single(msg)
	if bytes.Compare(pHash, rHash) != 0 {
		t.Error(errors.New("message did not decode correctly"))
	}
}

func BenchmarkSplit(b *testing.B) {
	msgSize := 1 << 16
	segSize := 1382
	var e error
	var payload []byte
	var hash sha256.Hash
	if payload, hash, e = testutils.GenerateTestMessage(msgSize); check(e) {
		b.Error(e)
	}
	_ = hash
	var sp, rp, Rp *prv.Key
	var sP, rP *pub.Key
	if sp, rp, sP, rP, e = testutils.GenerateTestKeyPairs(); check(e) {
		b.FailNow()
	}
	_, _, _ = sP, Rp, rp
	addr := address.FromPubKey(rP)
	for n := 0; n < b.N; n++ {
		params := packet.EP{
			To:     addr,
			From:   sp,
			Parity: 64,
			Data:   payload,
		}

		var splitted [][]byte
		if splitted, e = Split(params, segSize); check(e) {
			b.Error(e)
		}
		_ = splitted
	}
}

func TestRemovePacket(t *testing.T) {
	packets := make(packet.Packets, 10)
	for i := range packets {
		packets[i] = &packet.Packet{Seq: uint16(i)}
	}
	var seqs []uint16
	for i := range packets {
		seqs = append(seqs, packets[i].Seq)
	}
	discard := []int{1, 5, 6}
	for i := range discard {
		// Subtracting the iterator accounts for the backwards shift of
		// the shortened slice.
		packets = RemovePacket(packets, discard[i]-i)
	}
	var seqs2 []uint16
	for i := range packets {
		seqs2 = append(seqs2, packets[i].Seq)
	}
}
