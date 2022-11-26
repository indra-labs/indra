package segment

import (
	"bytes"
	"errors"
	"math/rand"
	"testing"

	"github.com/Indra-Labs/indra/pkg/key/address"
	"github.com/Indra-Labs/indra/pkg/key/prv"
	"github.com/Indra-Labs/indra/pkg/key/pub"
	"github.com/Indra-Labs/indra/pkg/packet"
	"github.com/Indra-Labs/indra/pkg/segcalc"
	"github.com/Indra-Labs/indra/pkg/sha256"
	"github.com/Indra-Labs/indra/pkg/testutils"
)

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
			segMap := segcalc.NewSegments(len(ep.Data), segSize, overhead, ep.Parity)
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
			var pkts packet.Packets
			var keys []*pub.Key
			for s := range splitted {
				var pkt *packet.Packet
				var from *pub.Key
				if _, from, e = packet.GetKeys(
					splitted[s]); e != nil {
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
			if bytes.Compare(pHash, rHash) != 0 {
				t.Error(errors.New("message did not decode" +
					" correctly"))
			}
		}
	}
}
