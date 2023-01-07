package packet

import (
	"errors"
	"math/rand"
	"testing"

	"github.com/indra-labs/indra/pkg/key/prv"
	"github.com/indra-labs/indra/pkg/key/pub"
	"github.com/indra-labs/indra/pkg/sha256"
	"github.com/indra-labs/indra/pkg/testutils"
)

func TestSplitJoin(t *testing.T) {
	msgSize := 2 << 19
	segSize := 1382
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
	addr := rP
	params := EP{
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
	var pkts Packets
	var keys []*pub.Key
	for i := range splitted {
		var pkt *Packet
		var from *pub.Key
		if from, e = GetKeys(splitted[i]); check(e) {
			log.I.Ln(i)
			continue
		}
		if pkt, e = Decode(splitted[i], from, rp); check(e) {
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
	if pHash != rHash {
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
	addr := rP
	for n := 0; n < b.N; n++ {
		params := EP{
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
	packets := make(Packets, 10)
	for i := range packets {
		packets[i] = &Packet{Seq: uint16(i)}
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

func TestSplitJoinFEC(t *testing.T) {
	msgSize := 2 << 15
	segSize := 1382
	var e error
	var sp, rp, Rp *prv.Key
	var sP, rP, RP *pub.Key
	if sp, rp, sP, rP, e = testutils.GenerateTestKeyPairs(); check(e) {
		t.FailNow()
	}
	_, _, _, _ = sP, Rp, RP, rp
	var parity []int
	for i := 1; i < 255; i *= 2 {
		parity = append(parity, i)
	}
	for i := range parity {
		var payload []byte
		var pHash sha256.Hash

		if payload, pHash, e = testutils.GenerateTestMessage(msgSize); check(e) {
			t.FailNow()
		}
		var punctures []int
		// Generate a set of numbers of punctures starting from equal to
		// parity in a halving sequence to reduce the number but see it
		// function.
		for punc := parity[i]; punc > 0; punc /= 2 {
			punctures = append(punctures, punc)
		}
		// Reverse the ordering just because.
		for p := 0; p < len(punctures)/2; p++ {
			punctures[p], punctures[len(punctures)-p-1] =
				punctures[len(punctures)-p-1], punctures[p]
		}
		addr := rP
		for p := range punctures {
			var splitted [][]byte
			ep := EP{
				To:     addr,
				From:   sp,
				Parity: parity[i],
				Length: len(payload),
				Data:   payload,
			}
			if splitted, e = Split(ep, segSize); check(e) {
				t.FailNow()
			}
			overhead := ep.GetOverhead()
			segMap := NewSegments(len(ep.Data), segSize, overhead, ep.Parity)
			for segs := range segMap {
				start, end := segMap[segs].DStart, segMap[segs].PEnd
				cnt := end - start
				par := segMap[segs].PEnd - segMap[segs].DEnd
				a := make([][]byte, cnt)
				for ss := range a {
					a[ss] = splitted[start:end][ss]
				}
				rand.Seed(int64(punctures[p]))
				rand.Shuffle(cnt,
					func(i, j int) {
						a[i], a[j] = a[j], a[i]
					})
				puncture := punctures[p]
				if puncture > par {
					puncture = par
				}
				for n := 0; n < puncture; n++ {
					copy(a[n][:100], make([]byte, 10))
				}
			}
			var pkts Packets
			var keys []*pub.Key
			for s := range splitted {
				var pkt *Packet
				var from *pub.Key
				if from, e = GetKeys(
					splitted[s]); e != nil {
					// we are puncturing, they some will
					// fail to decode
					continue
				}
				if pkt, e = Decode(splitted[s],
					from, rp); check(e) {
					continue
				}
				pkts = append(pkts, pkt)
				keys = append(keys, from)
			}
			// check all keys are the same
			prev := keys[0]
			for _, k := range keys[1:] {
				if !prev.Equals(k) {
					t.Error(e)
				}
				prev = k
			}
			var msg []byte
			if msg, e = Join(pkts); check(e) {
				t.FailNow()
			}
			rHash := sha256.Single(msg)
			if pHash != rHash {
				t.Error(errors.New("message did not decode" +
					" correctly"))
			}
		}
	}
}
