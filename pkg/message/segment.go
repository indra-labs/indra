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
func Split(ep EP, segSize int) (packets [][]byte, e error) {
	overhead := ep.GetOverhead()
	segMap := NewSegments(len(ep.Data), segSize, overhead, ep.Parity)
	// log.I.Ln(len(ep.Data), segSize, overhead, ep.Parity)
	// log.I.S(segMap)
	// pad the last part
	sp := segMap[len(segMap)-1]
	padLen := sp.SLen - sp.Last
	ep.Data = append(ep.Data, slice.NoisePad(padLen)...)
	var s [][]byte
	var start, end int
	for _, sm := range segMap {
		// Add the data segments.
		for curs := 0; curs < sm.DEnd-sm.DStart; curs++ {
			start += sm.SLen
			end = start + sm.SLen
			s = append(s, ep.Data[start:end])
		}
		// Add the empty parity segments, if any.
		for curs := 0; curs < sm.PEnd-sm.DEnd; curs++ {
			s = append(s, make([]byte, sm.SLen))
		}
		// If there is redundancy and parity segments were added,
		// generate the parity codes.
		if ep.Parity > 0 {
			section := s[sm.DStart:sm.PEnd]
			var rs *reedsolomon.RS
			dLen := sm.DEnd - sm.DStart
			pLen := sm.PEnd - sm.DEnd
			if rs, e = reedsolomon.New(dLen, pLen); check(e) {
				return
			}
			if e = rs.Encode(section); check(e) {
				return
			}
		}
		// Now we have the data encoded, next to encode the packets from
		// each of these segments.
		length := []int{
			sm.DEnd - sm.DStart,
			sm.PEnd - sm.DEnd,
		}
		lastEl := []int{
			length[0] - 1,
			length[1] - 1,
		}
		var packet []byte
		data := ep.Data
		for i := 0; i < 2; i++ {
			for curs := 0; curs < length[i]; curs++ {
				if curs == lastEl[i] {
					ep.Pad = padLen
				}
				ep.Data = s[ep.Seq]
				if packet, e = Encode(ep); check(e) {
					return
				}
				packets = append(packets, packet)
				ep.Seq++
			}
			ep.Pad = 0
		}
		ep.Data = data
	}
	return
}

const ErrNotEnough = "too many missing, redundancy %d, lost %d, section %d"
const ErrDupe = "found duplicate packet, no redundancy, decoding failed"
const ErrLostNoRedundant = "no redundancy with %d lost of %d"
const ErrMismatch = "found disagreement about common data in segment %d of %d" +
	" in field %s"

// JoinOld a collection of Packets together.
//
// Every message has a unique sender key, so once a packet is decoded, the
// pub.Print is the key to identifying associated packets.
func JoinOld(pkts Packets) (msg []byte, e error) {
	switch len(pkts) {
	case 0:
		e = errors.New("empty packets")
		return
	case 1:
		if pkts[0].Length == 1 {
			msg = pkts[0].Data
			return
		}
	}
	// assemble a list based on the expected total to be, and place the
	// packets into sequential order.
	nSegs := int(pkts[0].Length)
	payloadLen := len(pkts[0].Data)
	packets := make(Packets, nSegs)
	totP := int(pkts[0].Parity)
	// If we have no redundancy, if any are lost, fail now.
	if len(pkts) != nSegs && totP == 0 {
		e = fmt.Errorf(ErrLostNoRedundant, nSegs-len(pkts), nSegs)
		return
	}
	for i := range pkts {
		seq := pkts[i].Seq
		if packets[seq] != nil {
			log.I.Ln("found duplicate packet")
			// if we have no redundancy this means we have lost one.
			if totP == 0 {
				e = fmt.Errorf(ErrDupe)
				return
			}
		} else {
			packets[seq] = pkts[i]
		}
	}
	var segments [][]byte
	if totP <= 0 {
		msg = joinPacketData(packets)
	} else {
		// If there is redundancy in the message, check count of lost
		// packets.
		sections := nSegs / 256
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
				nLeft = nSegs - sections*256
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
				if packets[cur] == nil {
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

func countLostAndFound() {

}

func joinPacketData(listPackets Packets) (msg []byte) {
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

// Join a collection of Packets together.
//
// Every message has a unique sender key, so once a packet is decoded, the
// pub.Print is the key to identifying associated packets. The message collector
// must group them, they must have common addressee, message length and
// parity settings.
func Join(packets Packets) (msg []byte, e error) {
	if len(packets) == 0 {
		e = errors.New("empty packets")
		return
	}
	// Check that the data that should be common to all packets is common.
	to := packets[0].To
	p := packets[0]
	lp := len(packets)
	tot, red := packets[0].Length, packets[0].Parity
	for i, p := range packets {
		if i == 0 {
			continue
		}
		if check(to.Equals(p.To)) {
			e = fmt.Errorf(ErrMismatch, i, lp, "to")
			return
		}
		if p.Length != tot {
			e = fmt.Errorf(ErrMismatch, i, lp, "length")
			return
		}
		if p.Parity != red {
			e = fmt.Errorf(ErrMismatch, i, lp, "parity")
			return
		}
	}
	// Construct the segment map.
	overhead := p.GetOverhead()
	// log.I.Ln(int(p.Length), len(p.Data)+overhead, overhead, int(p.Parity))
	segMap := NewSegments(int(p.Length), len(p.Data)+overhead, overhead,
		int(p.Parity))
	log.I.S(segMap)

	for _, sm := range segMap {

		_ = sm
	}
	return
}
