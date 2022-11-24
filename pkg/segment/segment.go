package segment

import (
	"errors"
	"fmt"
	"sort"

	"github.com/Indra-Labs/indra"
	"github.com/Indra-Labs/indra/pkg/blake3"
	"github.com/Indra-Labs/indra/pkg/packet"
	"github.com/Indra-Labs/indra/pkg/segcalc"
	"github.com/Indra-Labs/indra/pkg/slice"
	log2 "github.com/cybriq/proc/pkg/log"
	"github.com/templexxx/reedsolomon"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

const ErrEmptyBytes = "cannot encode empty bytes"

// Split creates a series of packets including the defined Reed Solomon
// parameters for extra parity shards and the return encryption public key for a
// reply.
//
// The last packet that falls short of the segmentSize is padded random bytes.
//
// The segmentSize is inclusive of packet overhead plus the Seen key
// fingerprints at the end of the Packet.
func Split(ep packet.EP, segSize int) (packets [][]byte, e error) {
	if ep.Data == nil || len(ep.Data) == 0 {
		e = fmt.Errorf(ErrEmptyBytes)
		return
	}
	ep.Length = len(ep.Data)
	overhead := ep.GetOverhead()
	ss := segSize - overhead
	segments := slice.Segment(ep.Data, ss)
	segMap := segcalc.NewSegments(ep.Length, segSize, ep.GetOverhead(), ep.Parity)
	var p [][]byte
	p, e = segMap.AddParity(segments)
	for i := range p {
		ep.Data, ep.Seq = p[i], i
		var s []byte
		if s, e = packet.Encode(ep); check(e) {
			return
		}
		packets = append(packets, s)
	}
	return
}

const ErrDupe = "found duplicate packet, no redundancy, decoding failed"
const ErrLostNoRedundant = "no redundancy with %d lost of %d"
const ErrMismatch = "found disagreement about common data in segment %d of %d" +
	" in field %s"
const ErrNotEnough = "too many lost to recover in section %d, have %d, need " +
	"%d minimum"

// Join a collection of Packets together.
func Join(packets packet.Packets) (msg []byte, e error) {
	if len(packets) == 0 {
		e = errors.New("empty packets")
		return
	}
	// By sorting the packets we know we can iterate through them and detect
	// missing and duplicated items by simple rules.
	sort.Sort(packets)
	lp := len(packets)
	p := packets[0]
	// Construct the segments map.
	overhead := p.GetOverhead()
	segMap := segcalc.NewSegments(int(p.Length), len(p.Data)+overhead, overhead,
		int(p.Parity))
	segCount := segMap[len(segMap)-1].PEnd
	tot, red := p.Length, p.Parity
	prevSeq := p.Seq
	var discard []int
	// Check that the data that should be common to all packets is common,
	// and no sequence number is repeated.
	for i, ps := range packets {
		// Skip the first because we are comparing the rest to it. It is
		// arbitrary which item is reference because all should be the
		// same.
		if i == 0 {
			continue
		}
		// check that the sequence number isn't repeated.
		if ps.Seq == prevSeq {
			if red == 0 {
				e = fmt.Errorf(ErrDupe)
				return
			}
			// Check the data is the same, then discard the second
			// if they match.
			if blake3.Single(ps.Data).
				Equals(blake3.Single(packets[prevSeq].Data)) {

				discard = append(discard, int(ps.Seq))
				// No need to go on, we will discard this one.
				continue
			}
		}
		prevSeq = ps.Seq
		// All segments must specify the same total message length.
		if ps.Length != tot {
			e = fmt.Errorf(ErrMismatch, i, lp, "length")
			return
		}
		// All messages must have the same parity settings.
		if ps.Parity != red {
			e = fmt.Errorf(ErrMismatch, i, lp, "parity")
			return
		}
	}
	// Duplicates somehow found. Remove them.
	for i := range discard {
		// Subtracting the iterator accounts for the backwards shift of
		// the shortened slice.
		packets = RemovePacket(packets, discard[i]-i)
		lp--
	}
	// check there is all pieces if there is no redundancy.
	if red == 0 && lp < segCount {
		e = fmt.Errorf(ErrLostNoRedundant, segCount-lp, segCount)
		return
	}
	msg = make([]byte, 0, tot)
	// If all segments were received we can just concatenate the data shards
	if segCount == lp {
		for _, sm := range segMap {
			segments := make([][]byte, 0, sm.DEnd-sm.DStart)
			for i := sm.DStart; i < sm.DEnd; i++ {
				segments = append(segments, packets[i].Data)
			}
			msg = join(msg, segments, sm.SLen, sm.Last)
		}
		return
	}
	pkts := make(packet.Packets, segCount)
	// Collate to correctly ordered, so gaps are easy to find
	for i := range packets {
		pkts[packets[i].Seq] = packets[i]
	}
	// Count and collate found and lost segments, adding empty segments if
	// there is lost.
	for si, sm := range segMap {
		var lD, lP, hD, hP []int
		var segments [][]byte
		for i := sm.DStart; i < sm.DEnd; i++ {
			idx := i - sm.DStart
			if pkts[i] == nil {
				lD = append(lD, idx)
			} else {
				hD = append(hD, idx)
			}
		}
		for i := sm.DEnd; i < sm.PEnd; i++ {
			idx := i - sm.DStart
			if pkts[i] == nil {
				lP = append(lP, idx)
			} else {
				hP = append(hP, idx)
			}
		}
		dLen := sm.DEnd - sm.DStart
		lhD, lhP := len(hD), len(hP)
		if lhD+lhP < dLen {
			// segment cannot be corrected
			e = fmt.Errorf(ErrNotEnough, si, lhD+lhP, dLen)
			return
		}
		// if we have all the data segments we can just assemble and
		// return.
		if lhD == dLen {
			for i := sm.DStart; i < sm.DEnd; i++ {
				segments = append(segments, pkts[i].Data)
			}
			msg = join(msg, segments, sm.SLen, sm.Last)
			continue
		}
		// We have enough to do correction
		for i := sm.DStart; i < sm.PEnd; i++ {
			if pkts[i] == nil {
				segments = append(segments,
					make([]byte, sm.SLen))
			} else {
				segments = append(segments,
					pkts[i].Data)
			}
		}
		var rs *reedsolomon.RS
		if rs, e = reedsolomon.New(dLen, sm.PEnd-sm.DEnd); check(e) {
			return
		}
		have := append(hD, hP...)
		if e = rs.Reconst(segments, have, lD); check(e) {
			return
		}
		msg = join(msg, segments[:dLen], sm.SLen, sm.Last)
	}
	return
}

func join(msg []byte, segments [][]byte, sLen, last int) (b []byte) {
	b = slice.Cat(segments...)
	if sLen != last {
		b = b[:len(b)-sLen+last]
	}
	b = append(msg, b...)
	return
}

func RemovePacket(slice packet.Packets, s int) packet.Packets {
	return append(slice[:s], slice[s+1:]...)
}
