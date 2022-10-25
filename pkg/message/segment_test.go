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
	msgSize := 2 << 14
	segSize := 256
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
		To:         reciPub,
		From:       sendPriv,
		Blk:        blk1,
		Redundancy: 0,
		Seq:        0,
		Tot:        1,
		Data:       payload,
		Pad:        0,
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
	// log.I.Ln(len(pkts))
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
	_, _, _ = pHash, sendPub, msg
}

func TestSplitJoinFEC(t *testing.T) {
	msgSize := 2 << 16
	segSize := 256
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
		To:         reciPub,
		From:       sendPriv,
		Blk:        blk1,
		Redundancy: 1,
		Seq:        0,
		Tot:        1,
		Data:       payload,
		Pad:        0,
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
	// log.I.Ln(len(pkts))
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
	_, _, _ = pHash, sendPub, msg
}
