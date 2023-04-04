package engine

import (
	"errors"
	"testing"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/cloak"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/signer"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/util/tests"
)

func TestSplitJoin(t *testing.T) {
	log2.SetLogLevel(log2.Trace)
	segSize := 1024
	for sizeBase := 8; sizeBase < 24; sizeBase++ {
		msgSize := 1 << sizeBase
		log.D.Ln("size", msgSize)
		_, ks, _ := signer.New()
		var e error
		var payload []byte
		var pHash sha256.Hash
		_ = pHash
		if payload, pHash, e = tests.GenMessage(msgSize, "payload"); fails(e) {
			t.FailNow()
		}
		// log.D.S("original", payload)
		var sp, rp *prv.Key
		var rP, sP *pub.Key
		_ = sP
		if sp, rp, sP, rP, e = tests.GenerateTestKeyPairs(); fails(e) {
			t.FailNow()
		}
		addr := rP
		var splitted [][]byte
		params := Packet{
			ID:     nonce.NewID(),
			To:     addr,
			From:   sp,
			Length: uint32(len(payload)),
			Data:   payload,
			Parity: 0,
		}
		if splitted, e = Split(params, segSize, ks); fails(e) {
			t.Error(e)
			t.FailNow()
		}
		var pkts Packets
		var keys []*pub.Key
		for spl := range splitted {
			pkt := &Packet{}
			// log.D.S("prepacket", splitted[i])
			s := NewSpliceFrom(splitted[spl])
			if fails(pkt.Decode(s)) {
				t.Error("failed to decode packet")
				t.FailNow()
			}
			if !cloak.Match(pkt.CloakTo, rP.ToBytes()) {
				t.Error("failed to match cloak")
				t.FailNow()
			}
			if fails(pkt.Decrypt(rp, s)) {
				t.Error(e)
				t.FailNow()
			}
			// log.D.S("packet", pkt)
			pkts = append(pkts, pkt)
			keys = append(keys, pkt.fromPub)
		}
		var msg []byte
		if pkts, msg, e = JoinPackets(pkts); fails(e) {
			t.Error(e)
		}
		// log.D.S("msg", payload, msg)
		rHash := sha256.Single(msg)
		if pHash != rHash {
			t.Error(errors.New("message did not decode correctly"))
		}
	}
	
}
