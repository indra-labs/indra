package message

import (
	"crypto/rand"
	"testing"

	"github.com/Indra-Labs/indra/pkg/key/prv"
	"github.com/Indra-Labs/indra/pkg/key/pub"
	"github.com/Indra-Labs/indra/pkg/sha256"
)

//
// func TestSplitJoin(t *testing.T) {
// 	msgSize := 2 << 19
// 	segSize := 1472
// 	payload := make([]byte, msgSize)
// 	var e error
// 	var n int
// 	if n, e = rand.Read(payload); check(e) && n != msgSize {
// 		t.Error(e)
// 	}
// 	copy(payload, "payload")
// 	pHash := sha256.Single(payload)
// 	var sendPriv, reciPriv *prv.Key
// 	var sendPub, reciPub *pub.Key
// 	if sendPriv, e = prv.GenerateKey(); check(e) {
// 		t.Error(e)
// 	}
// 	sendPub = pub.Derive(sendPriv)
// 	if reciPriv, e = prv.GenerateKey(); check(e) {
// 		t.Error(e)
// 	}
// 	reciPub = pub.Derive(reciPriv)
// 	var blk1, blk2 cipher.Block
// 	if blk1, e = ciph.GetBlock(sendPriv, reciPub); check(e) {
// 		t.Error(e)
// 	}
// 	if blk2, e = ciph.GetBlock(reciPriv, sendPub); check(e) {
// 		t.Error(e)
// 	}
//
// 	params := EP{
// 		To:     reciPub,
// 		From:   sendPriv,
// 		Blk:    blk1,
// 		Length: len(payload),
// 		Data:   payload,
// 	}
// 	var splitted [][]byte
// 	if splitted, e = Split(params, segSize); check(e) {
// 		t.Error(e)
// 	}
// 	var pkts Packets
// 	var keys []*pub.Key
// 	for i := range splitted {
// 		var pkt *Packet
// 		var key *pub.Key
// 		if pkt, key, e = Decode(splitted[i]); check(e) {
// 			t.Error(e)
// 		}
// 		pkts = append(pkts, pkt.Decipher(blk2))
// 		keys = append(keys, key)
// 	}
// 	prev := keys[0]
// 	// check all keys are the same
// 	for _, k := range keys[1:] {
// 		if !prev.Equals(k) {
// 			t.Error(e)
// 		}
// 		prev = k
// 	}
// 	var msg []byte
// 	if msg, e = Join(pkts); check(e) {
// 		t.Error(e)
// 	}
// 	rHash := sha256.Single(msg)
// 	if bytes.Compare(pHash, rHash) != 0 {
// 		t.Error(errors.New("message did not decode correctly"))
// 	}
// }
//
// func TestSplitJoinFEC(t *testing.T) {
// 	msgSize := 2 << 17
// 	segSize := 1472
// 	var e error
// 	var sendPriv, reciPriv *prv.Key
// 	var reciPub *pub.Key
// 	_ = reciPriv
// 	var blk1, blk2 cipher.Block
// 	if sendPriv, reciPriv, _, reciPub, blk1, blk2, e =
// 		GenerateTestKeyPairs(); check(e) {
// 		t.FailNow()
// 	}
// 	var parity []int
// 	// parity = []int{
// 	// 	1,
// 	// 	4,
// 	// 	16,
// 	// 	64,
// 	// 	128,
// 	// }
//
// 	for i := 1; i < 255; i *= 2 {
// 		parity = append(parity, i)
// 	}
//
// 	for i := range parity {
// 		var payload []byte
// 		var pHash sha256.Hash
//
// 		if payload, pHash, e = GenerateTestMessage(msgSize); check(e) {
// 			t.FailNow()
// 		}
// 		var punctures []int
// 		// Generate a set of numbers of punctures starting from equal to
// 		// parity in a halving sequence to reduce the number but see it
// 		// function.
// 		for punc := parity[i]; punc > 0; punc /= 2 {
// 			punctures = append(punctures, punc)
// 		}
// 		// Reverse the ordering just because.
// 		for p := 0; p < len(punctures)/2; p++ {
// 			punctures[p], punctures[len(punctures)-p-1] =
// 				punctures[len(punctures)-p-1], punctures[p]
// 		}
// 		for p := range punctures {
// 			var splitted [][]byte
// 			ep := EP{
// 				To:     reciPub,
// 				From:   sendPriv,
// 				Blk:    blk1,
// 				Parity: parity[i],
// 				Length: len(payload),
// 				Data:   payload,
// 			}
// 			if splitted, e = Split(ep, segSize); check(e) {
// 				t.FailNow()
// 			}
// 			overhead := ep.GetOverhead()
// 			segMap := NewSegments(len(ep.Data), segSize, overhead, ep.Parity)
// 			for segs := range segMap {
// 				start, end := segMap[segs].DStart, segMap[segs].PEnd
// 				cnt := end - start
// 				par := segMap[segs].DStart - segMap[segs].DEnd
// 				a := make([][]byte, cnt)
// 				for ss := range a {
// 					a[ss] = splitted[start:end][ss]
// 				}
// 				mrand.Seed(int64(punctures[p]))
// 				mrand.Shuffle(cnt,
// 					func(i, j int) { a[i], a[j] = a[j], a[i] })
// 				puncture := punctures[p]
// 				if puncture > par {
// 					puncture = par
// 				}
// 				for n := 0; n < puncture; n++ {
// 					copy(a[n], make([]byte, 10))
// 				}
//
// 			}
// 			var pkts Packets
// 			var keys []*pub.Key
// 			for s := range splitted {
// 				var pkt *Packet
// 				var key *pub.Key
// 				if pkt, key, e = Decode(splitted[s]); check(e) {
// 					// we are puncturing, they some will fail to
// 					// decode
// 					continue
// 				}
// 				pkts = append(pkts, pkt.Decipher(blk2))
// 				keys = append(keys, key)
// 			}
//
// 			var msg []byte
// 			if msg, e = Join(pkts); check(e) {
// 				t.FailNow()
// 			}
// 			rHash := sha256.Single(msg)
// 			if bytes.Compare(pHash, rHash) != 0 {
// 				t.Error(errors.New("message did not decode" +
// 					" correctly"))
// 			}
// 		}
// 	}
// }

func TestSplit(t *testing.T) {
	msgSize := 2 << 19
	segSize := 1472
	parities := make([]int, 256)
	for i := 0; i < 256; i++ {
		parities[i] = i
	}
	var e error
	var payload []byte
	var hash sha256.Hash
	if payload, hash, e = GenerateTestMessage(msgSize); check(e) {
		t.Error(e)
	}
	_ = hash
	var sp, rp *prv.Key
	var sP, rP *pub.Key
	_, _ = rp, sP
	if sp, rp, sP, rP, e = GenerateTestKeyPairs(); check(e) {
		t.Error(e)
	}
	for i := range parities {
		params := EP{
			To:     rP,
			From:   sp,
			Parity: parities[i],
			Data:   payload,
		}

		var splitted [][]byte
		if splitted, e = Split(params, segSize); check(e) {
			t.Error(e)
		}
		_ = splitted
	}
}

// func BenchmarkSplit(b *testing.B) {
// 	msgSize := 2 << 16
// 	segSize := 4096
// 	payload := make([]byte, msgSize)
// 	var e error
// 	var n int
// 	if n, e = rand.Read(payload); check(e) && n != msgSize {
// 		b.Error(e)
// 	}
// 	copy(payload[:7], "payload")
// 	for n := 0; n < b.N; n++ {
// 		var sendPriv, reciPriv *prv.Key
// 		var reciPub *pub.Key
// 		if sendPriv, e = prv.GenerateKey(); check(e) {
// 			b.Error(e)
// 		}
// 		if reciPriv, e = prv.GenerateKey(); check(e) {
// 			b.Error(e)
// 		}
// 		reciPub = pub.Derive(reciPriv)
// 		var blk1 cipher.Block
// 		if blk1, e = ciph.GetBlock(sendPriv, reciPub); check(e) {
// 			b.Error(e)
// 		}
//
// 		params := EP{
// 			To:     reciPub,
// 			From:   sendPriv,
// 			Blk:    blk1,
// 			Parity: 64,
// 			Length: len(payload),
// 			Data:   payload,
// 		}
//
// 		var splitted [][]byte
// 		if splitted, e = Split(params, segSize); check(e) {
// 			b.Error(e)
// 		}
// 		_ = splitted
// 	}
// }

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

func GenerateTestKeyPairs() (sendPriv, reciPriv *prv.Key,
	sendPub, reciPub *pub.Key, e error) {
	if sendPriv, e = prv.GenerateKey(); check(e) {
		return
	}
	sendPub = pub.Derive(sendPriv)
	if reciPriv, e = prv.GenerateKey(); check(e) {
		return
	}
	reciPub = pub.Derive(reciPriv)
	// if blk1, e = ciph.GetBlock(sendPriv, reciPub); check(e) {
	// 	return
	// }
	// if blk2, e = ciph.GetBlock(reciPriv, sendPub); check(e) {
	// 	return
	// }
	return
}
