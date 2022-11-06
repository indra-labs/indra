package message

import (
	"bytes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"testing"

	"github.com/Indra-Labs/indra/pkg/ciph"
	"github.com/Indra-Labs/indra/pkg/key/prv"
	"github.com/Indra-Labs/indra/pkg/key/pub"
	"github.com/Indra-Labs/indra/pkg/sha256"
)

func TestSplitJoin(t *testing.T) {
	msgSize := 2 << 19
	segSize := 1472
	payload := make([]byte, msgSize)
	var e error
	var n int
	if n, e = rand.Read(payload); check(e) && n != msgSize {
		t.Error(e)
	}
	copy(payload, "payload")
	pHash := sha256.Single(payload)
	var sendPriv, reciPriv *prv.Key
	var sendPub, reciPub *pub.Key
	if sendPriv, e = prv.GenerateKey(); check(e) {
		t.Error(e)
	}
	sendPub = pub.Derive(sendPriv)
	if reciPriv, e = prv.GenerateKey(); check(e) {
		t.Error(e)
	}
	reciPub = pub.Derive(reciPriv)
	var blk1, blk2 cipher.Block
	if blk1, e = ciph.GetBlock(sendPriv, reciPub); check(e) {
		t.Error(e)
	}
	if blk2, e = ciph.GetBlock(reciPriv, sendPub); check(e) {
		t.Error(e)
	}

	params := EP{
		To:     reciPub,
		From:   sendPriv,
		Blk:    blk1,
		Parity: 0,
		Seq:    0,
		Length: len(payload),
		Data:   payload,
		Pad:    0,
	}
	var splitted [][]byte
	if splitted, e = Split(params, segSize); check(e) {
		t.Error(e)
	}
	var pkts Packets
	var keys []*pub.Key
	for i := range splitted {
		var pkt *Packet
		var key *pub.Key
		if pkt, key, e = Decode(splitted[i]); check(e) {
			t.Error(e)
		}
		pkts = append(pkts, pkt.Decipher(blk2))
		keys = append(keys, key)
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
	// rHash :=
	// _, _, _ = pHash, sendPub, msg
}

func TestSplitJoinFEC(t *testing.T) {
	msgSize := 2 << 18
	segSize := 1472
	var e error
	var payload []byte
	var pHash sha256.Hash

	if payload, pHash, e = GenerateTestMessage(msgSize); check(e) {
		t.FailNow()
	}
	var sendPriv, reciPriv *prv.Key
	var reciPub *pub.Key
	_ = reciPriv
	var blk1, blk2 cipher.Block
	if sendPriv, reciPriv, _, reciPub, blk1, blk2, e =
		GenerateTestKeyPairs(); check(e) {
		t.FailNow()
	}
	var splitted [][]byte
	parity := []int{
		1,
		2,
		4,
		8,
		16,
		32,
		64,
		128,
		160,
		192,
		216,
		240,
		248,
		252,
		255,
	}
	for i := range parity {

		if splitted, e = Split(EP{
			To:     reciPub,
			From:   sendPriv,
			Blk:    blk1,
			Parity: parity[i],
			Seq:    0,
			Length: len(payload),
			Data:   payload,
			Pad:    0,
		}, segSize); check(e) {
			t.FailNow()
		}
		var pkts Packets
		var keys []*pub.Key
		for i := range splitted {
			var pkt *Packet
			var key *pub.Key
			if pkt, key, e = Decode(splitted[i]); check(e) {
				// we are puncturing, they some will fail to
				// decode
			}
			pkts = append(pkts, pkt.Decipher(blk2))
			keys = append(keys, key)
		}
		// log.I.Ln(len(pkts))
		prev := keys[0]
		// check all keys are the same
		for _, k := range keys[1:] {
			if !prev.Equals(k) {
				t.FailNow()
			}
			prev = k
		}
		var msg []byte
		if msg, e = Join(pkts); check(e) {
			t.FailNow()
		}
		rHash := sha256.Single(msg)
		log.I.Ln(len(payload), len(msg))
		if bytes.Compare(pHash, rHash) != 0 {
			t.Error(errors.New("message did not decode correctly"))
		}
	}
}

func TestSplit(t *testing.T) {
	msgSize := 2 << 16
	segSize := 4096 // + Overhead
	payload := make([]byte, msgSize)
	var e error
	var n int
	if n, e = rand.Read(payload); check(e) && n != msgSize {
		t.Error(e)
	}
	copy(payload[:7], "payload")
	var sendPriv, reciPriv *prv.Key
	var reciPub *pub.Key
	if sendPriv, e = prv.GenerateKey(); check(e) {
		t.Error(e)
	}
	// sendPub = pub.Derive(sendPriv)
	if reciPriv, e = prv.GenerateKey(); check(e) {
		t.Error(e)
	}
	reciPub = pub.Derive(reciPriv)
	var blk1 cipher.Block
	if blk1, e = ciph.GetBlock(sendPriv, reciPub); check(e) {
		t.Error(e)
	}

	params := EP{
		To:     reciPub,
		From:   sendPriv,
		Blk:    blk1,
		Parity: 96,
		Seq:    0,
		Length: len(payload),
		Data:   payload,
		Pad:    0,
	}

	var splitted [][]byte
	if splitted, e = Split(params, segSize); check(e) {
		t.Error(e)
	}
	_ = splitted
}

func BenchmarkSplit(b *testing.B) {
	msgSize := 2 << 16
	segSize := 4096
	payload := make([]byte, msgSize)
	var e error
	var n int
	if n, e = rand.Read(payload); check(e) && n != msgSize {
		b.Error(e)
	}
	copy(payload[:7], "payload")
	for n := 0; n < b.N; n++ {
		var sendPriv, reciPriv *prv.Key
		var reciPub *pub.Key
		if sendPriv, e = prv.GenerateKey(); check(e) {
			b.Error(e)
		}
		// sendPub = pub.Derive(sendPriv)
		if reciPriv, e = prv.GenerateKey(); check(e) {
			b.Error(e)
		}
		reciPub = pub.Derive(reciPriv)
		var blk1 cipher.Block
		if blk1, e = ciph.GetBlock(sendPriv, reciPub); check(e) {
			b.Error(e)
		}

		params := EP{
			To:     reciPub,
			From:   sendPriv,
			Blk:    blk1,
			Parity: 64,
			Seq:    0,
			Length: len(payload),
			Data:   payload,
			Pad:    0,
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
	log.I.Ln(seqs)
	discard := []int{1, 5, 6}
	log.I.Ln("discarding", discard)
	for i := range discard {
		// Subtracting the iterator accounts for the backwards shift of
		// the shortened slice.
		packets = RemovePacket(packets, discard[i]-i)
	}
	var seqs2 []uint16
	for i := range packets {
		seqs2 = append(seqs2, packets[i].Seq)
	}
	log.I.Ln(seqs2)
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
	sendPub, reciPub *pub.Key, blk1, blk2 cipher.Block, e error) {
	if sendPriv, e = prv.GenerateKey(); check(e) {
		return
	}
	sendPub = pub.Derive(sendPriv)
	if reciPriv, e = prv.GenerateKey(); check(e) {
		return
	}
	reciPub = pub.Derive(reciPriv)
	if blk1, e = ciph.GetBlock(sendPriv, reciPub); check(e) {
		return
	}
	if blk2, e = ciph.GetBlock(reciPriv, sendPub); check(e) {
		return
	}
	return
}
