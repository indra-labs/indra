package packet

import (
	"errors"
	"git-indra.lan/indra-labs/indra/pkg/crypto"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/util/tests"
	"math/rand"
	"testing"
)

func BenchmarkSplit(b *testing.B) {
	msgSize := 1 << 16
	segSize := 1382
	var e error
	var payload []byte
	if payload, _, e = tests.GenMessage(msgSize, ""); fails(e) {
		b.Error(e)
	}
	var sp *crypto.Prv
	var rP *crypto.Pub
	if sp, _, _, rP, e = crypto.GenerateTestKeyPairs(); fails(e) {
		b.FailNow()
	}
	addr := rP
	for n := 0; n < b.N; n++ {
		params := &PacketParams{
			To:     addr,
			From:   sp,
			Parity: 64,
			Data:   payload,
		}
		var splitted [][]byte
		_, ks, _ := crypto.NewSigner()
		if _, splitted, e = SplitToPackets(params, segSize, ks); fails(e) {
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

func TestSplitJoin(t *testing.T) {
	log2.SetLogLevel(log2.Debug)
	msgSize := 1 << 19
	segSize := 1382
	var e error
	var payload []byte
	var pHash sha256.Hash
	if payload, pHash, e = tests.GenMessage(msgSize, ""); fails(e) {
		t.FailNow()
	}
	var sp, rp *crypto.Prv
	var rP *crypto.Pub
	if sp, rp, _, rP, e = crypto.GenerateTestKeyPairs(); fails(e) {
		t.FailNow()
	}
	addr := rP
	params := &PacketParams{
		To:     addr,
		From:   sp,
		Length: len(payload),
		Data:   payload,
		Parity: 128,
	}
	var splitted [][]byte
	_, ks, _ := crypto.NewSigner()
	if _, splitted, e = SplitToPackets(params, segSize, ks); fails(e) {
		t.Error(e)
	}
	var pkts Packets
	var keys []*crypto.Pub
	for i := range splitted {
		var pkt *Packet
		var from *crypto.Pub
		var to crypto.PubKey
		_ = to
		var iv nonce.IV
		if from, to, iv, e = GetKeysFromPacket(splitted[i]); fails(e) {
			log.I.Ln(i)
			continue
		}
		if !crypto.Match(to, rP.ToBytes()) {
			t.Error("did not match cloaked receiver key")
			t.FailNow()
		}
		if pkt, e = DecodePacket(splitted[i], from, rp, iv); fails(e) {
			t.Error(e)
		}
		pkts = append(pkts, pkt)
		keys = append(keys, from)
	}
	var msg []byte
	if pkts, msg, e = JoinPackets(pkts); fails(e) {
		t.Error(e)
	}
	rHash := sha256.Single(msg)
	if pHash != rHash {
		t.Error(errors.New("message did not decode correctly"))
	}
}

func TestSplitJoinFEC(t *testing.T) {
	log2.SetLogLevel(log2.Debug)
	msgSize := 1 << 18
	segSize := 1382
	var e error
	var sp, rp, Rp *crypto.Prv
	var sP, rP, RP *crypto.Pub
	if sp, rp, sP, rP, e = crypto.GenerateTestKeyPairs(); fails(e) {
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
		_, ks, _ := crypto.NewSigner()
		for p := range punctures {
			var splitted [][]byte
			ep := &PacketParams{
				To:     addr,
				From:   sp,
				Parity: parity[i],
				Length: len(payload),
				Data:   payload,
			}
			var ds int
			if ds, splitted, e = SplitToPackets(ep, segSize, ks); fails(e) {
				t.Error(e)
				t.FailNow()
			}
			log.D.Ln(ds, len(splitted))
			overhead := ep.GetOverhead()
			segMap := NewPacketSegments(ep.Length, segSize, overhead,
				ep.Parity)
			log.D.Ln(len(payload), len(splitted))
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
			var keys []*crypto.Pub
			for s := range splitted {
				var pkt *Packet
				var from *crypto.Pub
				var to crypto.PubKey
				_ = to
				var iv nonce.IV
				if from, to, iv, e = GetKeysFromPacket(
					splitted[s]); e != nil {
					// we are puncturing, they some will
					// fail to decode
					continue
				}
				if pkt, e = DecodePacket(splitted[s],
					from, rp, iv); fails(e) {
					continue
				}
				pkts = append(pkts, pkt)
				keys = append(keys, from)
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
