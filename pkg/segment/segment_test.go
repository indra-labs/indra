package segment

import (
	"bytes"
	"crypto/rand"
	"errors"
	mrand "math/rand"
	"testing"

	"github.com/Indra-Labs/indra/pkg/key/address"
	"github.com/Indra-Labs/indra/pkg/key/prv"
	"github.com/Indra-Labs/indra/pkg/key/pub"
	"github.com/Indra-Labs/indra/pkg/packet"
	"github.com/Indra-Labs/indra/pkg/sha256"
)

func TestSplitJoin(t *testing.T) {
	msgSize := 2 << 19
	segSize := 1472
	var e error
	var payload []byte
	var pHash sha256.Hash

	if payload, pHash, e = GenerateTestMessage(msgSize); check(e) {
		t.FailNow()
	}
	var sp, rp, Rp *prv.Key
	var sP, rP, RP *pub.Key
	if sp, rp, sP, rP, e = GenerateTestKeyPairs(); check(e) {
		t.FailNow()
	}
	_, _, _, _ = sP, Rp, RP, rp
	addr := address.FromPubKey(rP)
	params := packet.EP{
		To:     addr,
		From:   sp,
		Length: len(payload),
		Data:   payload,
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

func TestSplitJoinFEC(t *testing.T) {
	msgSize := 2 << 18
	segSize := 1472
	var e error
	var sp, rp, Rp *prv.Key
	var sP, rP, RP *pub.Key
	if sp, rp, sP, rP, e = GenerateTestKeyPairs(); check(e) {
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

		if payload, pHash, e = GenerateTestMessage(msgSize); check(e) {
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
		addr := address.FromPubKey(rP)
		for p := range punctures {
			var splitted [][]byte
			ep := packet.EP{
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
				mrand.Seed(int64(punctures[p]))
				mrand.Shuffle(cnt,
					func(i, j int) { a[i], a[j] = a[j], a[i] })
				puncture := punctures[p]
				if puncture > par {
					puncture = par
				}
				for n := 0; n < puncture; n++ {
					copy(a[n], make([]byte, 10))
				}

			}
			var pkts packet.Packets
			var keys []*pub.Key
			for s := range splitted {
				var pkt *packet.Packet
				var from *pub.Key
				if _, from, e = packet.GetKeys(splitted[s]); e != nil {
					// we are puncturing, they some will
					// fail to decode
					continue
				}
				if pkt, e = packet.Decode(splitted[s],
					from, rp); check(e) {
					continue
				}
				pkts = append(pkts, pkt)
				keys = append(keys, from)
			}

			var msg []byte
			if msg, e = Join(pkts); check(e) {
				t.FailNow()
			}
			rHash := sha256.Single(msg)
			if bytes.Compare(pHash, rHash) != 0 {
				t.Error(errors.New("message did not decode" +
					" correctly"))
			}
		}
	}
}

func BenchmarkSplit(b *testing.B) {
	msgSize := 2 << 16
	segSize := 1472
	var e error
	var payload []byte
	var hash sha256.Hash
	if payload, hash, e = GenerateTestMessage(msgSize); check(e) {
		b.Error(e)
	}
	_ = hash
	var sp, rp, Rp *prv.Key
	var sP, rP *pub.Key
	if sp, rp, sP, rP, e = GenerateTestKeyPairs(); check(e) {
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

func GenerateTestMessage(msgSize int) (msg []byte, hash sha256.Hash, e error) {
	msg = make([]byte, msgSize)
	var n int
	if n, e = rand.Read(msg); check(e) && n != msgSize {
		return
	}
	copy(msg, "payload")
	hash = sha256.Single(msg)
	return
}

func GenerateTestKeyPairs() (sp, rp *prv.Key, sP, rP *pub.Key, e error) {
	if sp, e = prv.GenerateKey(); check(e) {
		return
	}
	sP = pub.Derive(sp)
	if rp, e = prv.GenerateKey(); check(e) {
		return
	}
	rP = pub.Derive(rp)
	return
}
