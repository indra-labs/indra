package engine

import (
	"errors"
	"math/rand"
	"testing"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/cloak"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/signer"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/util/tests"
)

func TestSplitJoinFEC(t *testing.T) {
	log2.SetLogLevel(log2.Trace)
	_, ks, _ := signer.New()
	msgSize := 2 << 10
	segSize := 1382
	var e error
	var sp, rp, Rp *prv.Key
	var sP, rP, RP *pub.Key
	if sp, rp, sP, rP, e = tests.GenerateTestKeyPairs(); fails(e) {
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
		if payload, pHash, e = tests.GenMessage(msgSize, "b0rk"); fails(e) {
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
			params := Packet{
				To:     addr,
				From:   sp,
				Parity: byte(parity[i]),
				Length: uint32(len(payload)),
				Data:   payload,
			}
			if splitted, e = Split(params, segSize, ks); fails(e) {
				t.Error(e)
				t.FailNow()
			}
			overhead := PacketHeaderLen
			segMap := NewPacketSegments(len(params.Data), segSize, overhead,
				int(params.Parity))
			for segs := range segMap {
				start := segMap[segs].DStart
				end := segMap[segs].PEnd
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
			for spl := range splitted {
				pkt := &Packet{}
				// log.D.S("prepacket", splitted[i])
				s := NewSpliceFrom(splitted[spl])
				if fails(pkt.Decode(s)) {
					// we are puncturing, they some will
					// fail to decode
					continue
				}
				if !cloak.Match(pkt.CloakTo, rP.ToBytes()) {
					// we are puncturing, they some will
					// fail to decode
					continue
				}
				if fails(pkt.Decrypt(rp, s)) {
					// we are puncturing, they some will
					// fail to decode
					continue
				}
				pkts = append(pkts, pkt)
				keys = append(keys, pkt.fromPub)
			}
			var msg []byte
			if pkts, msg, e = JoinPackets(pkts); fails(e) {
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

func BenchmarkSplit(b *testing.B) {
	msgSize := 1 << 16
	segSize := 1382
	var e error
	var payload []byte
	if payload, _, e = tests.GenMessage(msgSize, ""); fails(e) {
		b.Error(e)
	}
	var sp *prv.Key
	var rP *pub.Key
	if sp, _, _, rP, e = tests.GenerateTestKeyPairs(); fails(e) {
		b.FailNow()
	}
	addr := rP
	_, ks, _ := signer.New()
	for n := 0; n < b.N; n++ {
		params := Packet{
			To:     addr,
			From:   sp,
			Parity: 64,
			Data:   payload,
		}
		
		var splitted [][]byte
		if splitted, e = Split(params, segSize, ks); fails(e) {
			b.Error(e)
		}
		_ = splitted
	}
	
	// Example benchmark results show about 10Mb/s/thread throughput
	// handling 64Kb messages.
	//
	// goos: linux
	// goarch: amd64
	// pkg: git-indra.lan/indra-labs/indra/pkg/packet
	// cpu: AMD Ryzen 7 5800H with Radeon Graphics
	// BenchmarkSplit
	// BenchmarkSplit-16    	     157	   7670080 ns/op
	// PASS
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
