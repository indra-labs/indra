package engine

import (
	"crypto/rand"
	"testing"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
	"git-indra.lan/indra-labs/indra/pkg/util/tests"
)

func TestEncode_Decode(t *testing.T) {
	log2.SetLogLevel(log2.Trace)
	msgSize := 256
	payload := make([]byte, msgSize)
	var e error
	var n int
	if n, e = rand.Read(payload); fails(e) && n != msgSize {
		t.Error(e)
		t.FailNow()
	}
	payload = append([]byte("payload"), payload...)
	var sp, rp *prv.Key
	var sP, rP *pub.Key
	if sp, rp, sP, rP, e = tests.GenerateTestKeyPairs(); fails(e) {
		t.Error(e)
		t.FailNow()
	}
	_, _ = rp, sP
	addr := rP
	pkt := &Packet{
		ID:     nonce.NewID(),
		To:     addr,
		From:   sp,
		Parity: 64,
		Length: uint32(len(payload)),
		Data:   payload,
	}
	log.D.S("packet", pkt)
	pl := pkt.Len()
	s := NewSplice(pl)
	if e = pkt.Encode(s); fails(e) {
		t.Error(e)
		t.FailNow()
	}
	log.D.S("pkt", pkt)
	b := make(slice.Bytes, s.Len())
	// copying is what a network does after all.
	copy(b, s.GetAll())
	s1 := NewSpliceFrom(b)
	pkt1 := &Packet{}
	if e = pkt1.Decode(s1); fails(e) {
		t.Error(e)
		t.FailNow()
	}
	if e = pkt1.Decrypt(rp, s1); fails(e) {
		t.Error(e)
		t.FailNow()
	}
	log.D.S("decryptid", pkt1)
	if string(pkt.Data) != string(pkt1.Data) {
		t.FailNow()
	}
}
