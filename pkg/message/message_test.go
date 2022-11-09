package message

import (
	"bytes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	mrand "math/rand"
	"testing"

	"github.com/Indra-Labs/indra/pkg/ciph"
	"github.com/Indra-Labs/indra/pkg/key/prv"
	"github.com/Indra-Labs/indra/pkg/key/pub"
	"github.com/Indra-Labs/indra/pkg/sha256"
)

func TestEncode_Decode(t *testing.T) {
	msgSize := mrand.Intn(3072) + 1024
	payload := make([]byte, msgSize)
	var e error
	var n int
	if n, e = rand.Read(payload); check(e) && n != msgSize {
		t.Error(e)
	}
	payload = append([]byte("payload"), payload...)
	pHash := sha256.Single(payload)
	var sendPriv, reciPriv *prv.Key
	var sendPub, reciPub *pub.Key
	if sendPriv, e = prv.GenerateKey(); check(e) {
		t.Error(e)
	}
	sendPub = pub.Derive(sendPriv)
	_ = sendPub
	if reciPriv, e = prv.GenerateKey(); check(e) {
		t.Error(e)
	}
	reciPub = pub.Derive(reciPriv)
	var blk cipher.Block
	if blk, e = ciph.GetBlock(sendPriv, reciPub); check(e) {
		t.Error(e)
	}
	var pkt []byte
	params := EP{
		To:   reciPub,
		From: sendPriv,
		Blk:  blk,
		Data: payload,
	}
	if pkt, e = Encode(params); check(e) {
		t.Error(e)
	}
	var f *Packet
	if f, e = Decode(pkt, blk); check(e) {
		t.Error(e)
	}
	dHash := sha256.Single(f.Data)
	if bytes.Compare(pHash, dHash) != 0 {
		t.Error(errors.New("encode/decode unsuccessful"))
	}

}
