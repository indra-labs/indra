package message

import (
	"errors"
	"fmt"
	"sort"

	"github.com/Indra-Labs/indra/pkg/sha256"
	"github.com/Indra-Labs/indra/pkg/slice"
	"github.com/templexxx/reedsolomon"
)

const ErrEmptyBytes = "cannot encode empty bytes"

// Split creates a series of packets including the defined Reed Solomon
// parameters for extra parity shards.
//
// The last packet that falls short of the segmentSize is padded random bytes.
//
// The segmentSize is inclusive of packet overhead plus the Seen key
// fingerprints at the end of the Packet.
func Split(ep EP, segSize int) (packets [][]byte, e error) {
	if ep.Data == nil || len(ep.Data) == 0 {
		e = fmt.Errorf(ErrEmptyBytes)
		return
	}
	overhead := ep.GetOverhead()
	segMap := NewSegments(len(ep.Data), segSize, overhead, ep.Parity)
	// pad the last part
	sp := segMap[len(segMap)-1]
	padLen := sp.SLen - sp.Last
	ep.Data = append(ep.Data, slice.NoisePad(padLen)...)
	var s [][]byte
	var start, end int
	for i, sm := range segMap {
		// Add the data segments.
		for curs := 0; curs < sm.DEnd-sm.DStart; curs++ {
			end = start + sm.SLen
			if i == sm.DEnd-sm.DStart {
				log.I.Ln("last", sm.Last)
				end = start + sm.Last
			}
			s = append(s, ep.Data[start:end])
			start += sm.SLen
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
		for j := 0; j < 2; j++ {
			for curs := 0; curs < length[j]; curs++ {
				if curs == lastEl[j] {
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

const ErrDupe = "found duplicate packet, no redundancy, decoding failed"
const ErrLostNoRedundant = "no redundancy with %d lost of %d"
const ErrMismatch = "found disagreement about common data in segment %d of %d" +
	" in field %s"

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
	// By sorting the packets we know we can iterate through them and detect
	// missing and duplicated items by simple rules.
	sort.Sort(packets)
	lp := len(packets)
	p := packets[0]
	// Construct the segments map.
	overhead := p.GetOverhead()
	segMap := NewSegments(int(p.Length), len(p.Data)+overhead, overhead,
		int(p.Parity))
	// check there is all pieces if there is no redundancy.
	segCount := segMap[len(segMap)-1].PEnd
	tot, red := p.Length, p.Parity
	if red == 0 && lp < segCount {
		e = fmt.Errorf(ErrLostNoRedundant, segCount-lp, segCount)
		return
	}
	to := p.To
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
			if sha256.Single(ps.Data).
				Equals(sha256.Single(packets[prevSeq].Data)) {

				discard = append(discard, int(ps.Seq))
				// No need to go on, we will discard this one.
				continue
			}
		}
		prevSeq = ps.Seq
		// All segments must be addressed to the same key.
		if check(to.Equals(ps.To)) {
			e = fmt.Errorf(ErrMismatch, i, lp, "to")
			return
		}
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
	// If redundancy is zero, and we have the expected amount, we can just
	// join them and return.
	if segMap[len(segMap)-1].PEnd == len(packets) {
		for _, sm := range segMap {
			var segment [][]byte
			for i := sm.DStart; i < sm.DEnd; i++ {
				segment = append(segment, packets[i].Data)
			}
			msg = append(msg, slice.Concatenate(segment...)...)
			// Slice off the padding if any.
			msg = msg[:len(msg)-sm.SLen+sm.Last]
		}
		return
	}
	// Make a new list of the data segments to pin our items to, which will
	// be bigger to the degree of segments lost.
	listPackets := make(Packets, segCount)
	for i := range packets {
		listPackets[packets[i].Seq] = packets[i]
	}
	// Collate the sections and fill up have/lost lists.
	for _, sm := range segMap {
		var segments [][]byte
		var haveD, haveP, lost []int
		start, parity, end := sm.DStart, sm.DEnd, sm.PEnd
		sLen := sm.SLen
		for i := start; i < parity; i++ {
			if listPackets[i] == nil {
				lost = append(lost, i)
				segments = append(segments, make([]byte, sLen))
			} else {
				haveD = append(haveD, i)
				segments = append(segments, listPackets[i].Data)
			}
		}
		for i := parity; i < end; i++ {
			if listPackets[i] == nil {
				lost = append(lost, i)
				segments = append(segments, make([]byte, sLen))
			} else {
				haveP = append(haveP, i)
				segments = append(segments, listPackets[i].Data)
			}
		}
		dLen := parity - start
		if len(haveD) == dLen {
			msg = append(msg, slice.Concatenate(segments[:dLen]...)...)
			msg = msg[:len(msg)-sm.SLen+sm.Last]
			continue
		}
		pLen := end - parity
		var rs *reedsolomon.RS
		if rs, e = reedsolomon.New(dLen, pLen); check(e) {
			return
		}
		if e = rs.Reconst(
			segments, append(haveD, haveP...), lost); check(e) {
			return
		}
		msg = append(msg, slice.Concatenate(segments[:dLen]...)...)
		msg = msg[:len(msg)-sm.SLen+sm.Last]
	}
	return
}

func RemovePacket(slice Packets, s int) Packets {
	return append(slice[:s], slice[s+1:]...)
}
