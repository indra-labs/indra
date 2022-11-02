package message

import (
	"errors"
	"fmt"

	"github.com/Indra-Labs/indra/pkg/slice"
	"github.com/templexxx/reedsolomon"
)

// Split creates a series of packets including the defined Reed Solomon
// parameters for extra parity shards.
//
// The last packet that falls short of the segmentSize is padded random bytes.
//
// The segmentSize is inclusive of packet overhead plus the Seen key
// fingerprints at the end of the Packet.
func Split(ep EP, segSize int) (pkts [][]byte, e error) {
	overhead := ep.GetOverhead()
	segMap := NewSegments(len(ep.Data), segSize, overhead, ep.Redundancy)
	// pad the last part
	sp := segMap[len(segMap)-1]
	ep.Data = append(ep.Data, slice.NoisePad(sp.SLen-sp.Last)...)
	var s [][]byte
	var start, end int
	for _, sm := range segMap {
		for curs := 0; curs < sm.DEnd-sm.DStart; curs++ {
			start = curs * sm.SLen
			end = start + sm.SLen
			s = append(s, ep.Data[start:end])
		}
		for curs := 0; curs < sm.PEnd-sm.DEnd; curs++ {
			start = curs * sm.SLen
			end = start + sm.SLen
			s = append(s, make([]byte, end-start))
		}
	}
	if ep.Redundancy > 0 {
		for _, sm := range segMap {
			section := s[sm.DStart:sm.PEnd]
			var rs *reedsolomon.RS
			if rs, e = reedsolomon.New(sm.DEnd-sm.DStart,
				sm.PEnd-sm.DEnd); check(e) {
				return
			}
			if e = rs.Encode(section); check(e) {
				return
			}
		}
	}
	// Now we have the data encoded, next to encode the packets from each of
	// these segments.
	ep.Tot = len(s)
	for _, sm := range segMap {
		lastEl := sm.DEnd - sm.DStart - 1
		for curs := 0; curs < sm.DEnd-sm.DStart; curs++ {
			if curs == lastEl {
				ep.Pad = sm.SLen - sm.Last
			}
			ep.Data = s[ep.Seq]
			var pkt []byte
			if pkt, e = Encode(ep); check(e) {
				return
			}
			pkts = append(pkts, pkt)
			ep.Seq++

		}
		ep.Pad = 0
		lastEl = sm.DEnd - sm.DStart - 1
		for curs := 0; curs < sm.PEnd-sm.DEnd; curs++ {
			if curs == lastEl {
				ep.Pad = sm.SLen - sm.Last
			}
			ep.Data = s[ep.Seq]
			var pkt []byte
			if pkt, e = Encode(ep); check(e) {
				return
			}
			pkts = append(pkts, pkt)
			ep.Seq++
		}
	}
	return
}

const ErrNotEnough = "too many missing, redundancy %d, lost %d, section %d"
const ErrDupe = "found duplicate packet, no redundancy, decoding failed"
const ErrLostNoRedundant = "no redundancy with %d lost of %d"

// Join a collection of Packets together.
//
// Every message has a unique sender key, so once a packet is decoded, the
// pub.Print is the key to identifying associated packets.
func Join(pkts Packets) (msg []byte, e error) {
	switch len(pkts) {
	case 0:
		e = errors.New("empty packets")
		return
	case 1:
		if pkts[0].Tot == 1 {
			msg = pkts[0].Data
			return
		}
	}
	// assemble a list based on the expected total to be, and place the
	// packets into sequential order.
	msgSegLen := int(pkts[0].Tot)
	payloadLen := len(pkts[0].Data)
	listPackets := make(Packets, msgSegLen)
	totP := int(pkts[0].Redundancy)
	// If we have no redundancy, if any are lost, fail now.
	if len(pkts) != msgSegLen && totP == 0 {
		e = fmt.Errorf(ErrLostNoRedundant,
			msgSegLen-len(pkts), msgSegLen)
		return
	}
	for i := range pkts {
		seq := pkts[i].Seq
		if listPackets[seq] != nil {
			log.I.Ln("found duplicate packet")
			// if we have no redundancy this means we have lost one.
			if totP == 0 {
				e = fmt.Errorf(ErrDupe)
				return
			}
		} else {
			listPackets[seq] = pkts[i]
		}
	}
	var segments [][]byte
	if totP <= 0 {
		// In theory, we have all segments, as packet decoding was
		// evidently successful, thus all checksums passed, and the
		// concatenation of the packets should also be successful.
		var totalLen int
		for i := range listPackets {
			totalLen += len(listPackets[i].Data)
		}
		// Pre-allocate the space.
		msg = make([]byte, 0, totalLen)
		for i := range listPackets {
			msg = append(msg, listPackets[i].Data...)
		}
	} else {
		// If there is redundancy in the message, check count of lost
		// packets.
		sections := msgSegLen / 256
		var counter, start, end int
		var nLost []int
		// In each section, a maximum of the number equal to the parity
		// segments can be lost before recovery of the message becomes
		// impossible, or the same ratio combined with the last
		// remaining section (last section can be shorter, last segment
		// of section can be shorter than the rest of the segments).
		totD := 256 - totP
		var short, sPad []int
		for i := 0; i <= sections; i++ {
			var lostSomeD bool
			nLeft := 256
			nLost = append(nLost, 0)
			if i == sections {
				nLeft = msgSegLen - sections*256
			}
			end += nLeft
			nTotP := nLeft * totP / totD
			nTotD := nLeft - nTotP
			log.I.F("nLeft %d nTotD %d nTotP %d, start %d end %d",
				nLeft, nTotD, nTotP, start, end)
			var cur int
			var lostD, haveDP []int
			n := i * 256
			for ; counter < nLeft; counter++ {
				cur = n + counter
				// count how many packets were lost
				if listPackets[cur] == nil {
					nLost[i]++
					if nLost[i] > nTotP {
						e = fmt.Errorf(ErrNotEnough,
							totP, nLost[i], i)
						return
					}
					if cur < nTotD {
						// we lost some data
						lostSomeD = true
						// Make note for rs.Reconst.
						// We don't need to recover
						// parity shards!
						lostD = append(lostD, cur)
					}
					// put empty segment in here
				} else {
					// make note of segments we have
					haveDP = append(haveDP, cur)
				}
			}
			if !lostSomeD {
				// No need for reconstruction, we have all the
				// data shards, add them to the recovered
				// segments.
				pktSect := pkts[n : n+totD]
				for _, p := range pktSect {
					segments = append(segments, p.Data)
				}
			} else {
				var s [][]byte
				s, short, sPad, e = reconst(pkts,
					n,
					nTotD,
					nTotP,
					payloadLen,
					haveDP,
					lostD)
				if check(e) {
					return
				}
				for i := range short {
					s[short[i]] =
						s[short[i]][:len(
							s[short[i]])-sPad[i]]
				}
				segments = append(segments, s...)
			}
			counter = 0
			start += nLeft
		}
		dataLen := slice.SumLen(segments...)
		msg = make([]byte, 0, dataLen)
		for i := range segments {
			msg = append(msg, segments[i]...)
		}
	}
	return
}

func reconst(pkts Packets, n, nTotD, nTotP, payloadLen int,
	haveDP, lostD []int) (segments [][]byte, short, sPad []int, e error) {

	var rs *reedsolomon.RS
	rs, e = reedsolomon.New(nTotD, nTotP)
	if check(e) {
		return
	}
	var shards [][]byte
	pktSect := pkts[n : n+nTotD+nTotP]
	for i, p := range pktSect {
		if p == nil {
			shards = append(shards, make([]byte, payloadLen))
		} else {
			data := p.Data
			if len(data) < payloadLen {
				short = append(short, i)
				pad := payloadLen - len(data)
				sPad = append(sPad, pad)
				p.Data = append(p.Data, make([]byte, pad)...)
			}
			shards = append(shards, p.Data)
		}
	}
	if e = rs.Reconst(shards, haveDP, lostD); check(e) {
		return
	}
	// we should now be able to add the section to
	// the segments
	segments = append(segments, shards[:nTotD]...)
	return
}
