package engine

import (
	"errors"
	"fmt"
	"sort"
	
	"github.com/templexxx/reedsolomon"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/signer"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

const (
	ErrEmptyBytes      = "cannot encode empty bytes"
	ErrDupe            = "found duplicate packet, no redundancy, decoding failed"
	ErrLostNoRedundant = "no redundancy with %d lost of %d"
	ErrMismatch        = "found disagreement about common data in segment %d of %d" +
		" in field %s value: got %v expected %v"
	ErrNotEnough = "too many lost to recover in section %d, have %d, need " +
		"%d minimum"
	ErrDifferentID = "different ID"
)

// Split creates a series of packets including the defined Reed Solomon
// parameters for extra parity shards and the return encryption public key for a
// reply.
func Split(pp Packet, segSize int, ks *signer.KeySet) (packets [][]byte, e error) {
	if pp.Data == nil || len(pp.Data) == 0 {
		e = fmt.Errorf(ErrEmptyBytes)
		return
	}
	pp.Length = uint32(len(pp.Data))
	overhead := PacketHeaderLen
	ss := segSize - overhead
	segments := slice.Segment(pp.Data, ss)
	segMap := NewPacketSegments(int(pp.Length), segSize, overhead, int(pp.Parity))
	log.D.Ln("segMap", segMap, int(pp.Length), segSize, overhead, int(pp.Parity))
	var pkts [][]byte
	pkts, e = segMap.AddParity(segments)
	for i := range pkts {
		pkt := &Packet{
			ID:     pp.ID,
			To:     pp.To,
			From:   ks.Next(),
			Seq:    uint16(i),
			Parity: pp.Parity,
			Length: pp.Length,
			Data:   pkts[i],
		}
		// log.D.S(pkt)
		s := NewSplice(pkt.Len())
		if e = pkt.Encode(s); fails(e) {
			return
		}
		packets = append(packets, s.GetAll())
	}
	return
}

// JoinPackets joins a collection of Packets together.
func JoinPackets(packets Packets) (pact Packets, msg []byte, e error) {
	pact = packets
	// We return this because cleaning shouldn't be done twice on old stale
	// data so failed joins still return the cleaned packet slice.
	if len(pact) == 0 {
		e = errors.New("empty packets")
		return
	}
	// By sorting the packets we know we can iterate through them and detect
	// missing and duplicated items by simple rules.
	sort.Sort(pact)
	lp := len(pact)
	p := pact[0]
	// Construct the segments map.
	overhead := PacketHeaderLen
	segMap := NewPacketSegments(
		int(p.Length), len(p.Data)+overhead, overhead, int(p.Parity))
	log.D.S("joinSegmap", segMap, int(p.Length), int(p.Length)+overhead,
		overhead,
		int(p.Parity))
	segCount := segMap[len(segMap)-1].PEnd
	length, red := p.Length, p.Parity
	log.D.Ln("length", length)
	id := p.ID
	prevSeq := p.Seq
	var discard []int
	// Check that the data that should be common to all pact is common, and
	// no sequence number is repeated.
	for i, ps := range pact {
		// Skip the first because we are comparing the rest to it. It is
		// arbitrary which item is reference because all should be the same.
		if i == 0 {
			continue
		}
		// fail that the sequence number isn't repeated.
		if ps.Seq == prevSeq {
			if red == 0 {
				e = fmt.Errorf(ErrDupe)
				return
			}
			// Check the data is the same, then discard the second if they
			// match.
			if sha256.Single(ps.Data) ==
				sha256.Single(pact[prevSeq].Data) {
				
				discard = append(discard, int(ps.Seq))
				// Node need to go on, we will discard this one.
				continue
			}
		}
		prevSeq = ps.Seq
		// All messages must have the same parity settings.
		if ps.Parity != red {
			e = fmt.Errorf(ErrMismatch, i, lp, "parity",
				ps.Parity, red)
			return
		}
		// All segments must specify the same total message length.
		if ps.Length != length {
			e = fmt.Errorf(ErrMismatch, i, lp, "length",
				ps.Length, length)
			return
		}
		// This should never happen but if it does the entire packet batch is
		// lost. Precautionary, since such an event is a bug and unsanitary
		// data.
		if ps.ID != id {
			e = fmt.Errorf(ErrDifferentID)
			return
		}
	}
	// Duplicates somehow found. Remove them.
	for i := range discard {
		// Subtracting the iterator accounts for the backwards shift of the
		// shortened slice.
		pact = RemovePacket(pact, discard[i]-i)
		lp--
	}
	log.D.Ln("length", length)
	// check there is all pieces if there is no redundancy.
	if red == 0 && lp < segCount {
		e = fmt.Errorf(ErrLostNoRedundant, segCount-lp, segCount)
		return
	}
	msg = make([]byte, 0, length)
	// If all segments were received we can just concatenate the data shards.
	if segCount == lp {
		log.T.Ln("all segments available")
		for _, sm := range segMap {
			segments := make([][]byte, 0, sm.DEnd-sm.DStart)
			for i := sm.DStart; i < sm.DEnd; i++ {
				segments = append(segments, pact[i].Data)
			}
			msg = join(msg, segments, sm.SLen, sm.Last)[:length]
			// log.D.S("length", length, msg)
		}
		return
	}
	log.D.Ln("length", length)
	// pact = make(Packets, len(pact))
	// Collate to correctly ordered, so gaps are easy to find
	// for i := range pact {
	// 	pact[pact[i].Seq] = pact[i]
	// }
	// for i := range pact {
	// 	log.D.S("pkts", i, pact[i].Seq, pact[i].Data)
	// 	log.D.Ln("pact", i)
	// }
	// Count and collate found and lost segments, adding empty segments if there
	// is lost.
	for si, sm := range segMap {
		dSegs := sm.DEnd - sm.DStart
		var lD, lP, hD, hP []int
		var segments [][]byte
		log.D.S("sm", sm)
		// log.D.S("pact", pkts)
		log.D.Ln("number data", dSegs)
		for i := sm.DStart; i < sm.DEnd; i++ {
			if i > len(pact)-1 {
				break
			}
			idx := i - sm.DStart
			if pact[i] == nil {
				lD = append(lD, idx)
			} else {
				hD = append(hD, idx)
			}
		}
		lhD, lhP := len(hD), len(hP)
		log.D.Ln("data lhD lhP pkts", lhD, lhP, len(pact))
		log.D.S("index", sm)
		for i := sm.DEnd; i < sm.PEnd; i++ {
			log.D.Ln("collating", i)
			if i > len(pact)-1 {
				break
			}
			idx := i - sm.DStart
			log.D.Ln("idx", i, idx, sm.DStart, sm.DEnd, sm.PEnd)
			if pact[i] == nil {
				lP = append(lP, idx)
			} else {
				hP = append(hP, idx)
			}
		}
		lhD, lhP = len(hD), len(hP)
		log.D.Ln("parity lhD lhP pkts", lhD, lhP, dSegs,
			len(pact))
		if lhD+lhP < dSegs {
			// segment cannot be corrected
			e = fmt.Errorf(ErrNotEnough, si, lhD+lhP, len(pact))
			return
		}
		// if we have all the data segments we can just assemble and return.
		if lhD == dSegs {
			log.D.Ln("wut", sm.DStart, sm.DEnd)
			log.D.S("smap", sm)
			for i := sm.DStart; i < sm.DEnd; i++ {
				segments = append(segments, pact[i].Data)
			}
			msg = join(msg, segments, sm.SLen, sm.Last)[:length]
			log.D.S("length", length, msg)
			continue
		}
		// We have enough to do correction
		for i := sm.DStart; i < sm.PEnd; i++ {
			if pact[i] == nil {
				segments = append(segments,
					make([]byte, sm.SLen))
			} else {
				segments = append(segments,
					pact[i].Data)
			}
		}
		var rs *reedsolomon.RS
		// if rs, e = reedsolomon.New(dLen, sm.PEnd-sm.DEnd); fails(e) {
		// 	return
		// }
		have := append(hD, hP...)
		if e = rs.Reconst(segments, have, lD); fails(e) {
			return
		}
		// msg = join(msg, segments[:dLen], sm.SLen, sm.Last)
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

func RemovePacket(slice Packets, s int) Packets {
	return append(slice[:s], slice[s+1:]...)
}
