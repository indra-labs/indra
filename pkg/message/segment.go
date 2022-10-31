package message

import (
	"errors"
	"sort"

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
	// Now we have the data encoded, next to encode the packets from each
	// of these segments.
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
		msg = pkts[0].Payload
		return
	}
	sort.Sort(pkts)

	e = errors.New("not yet implemented data shard recovery")
	return
}
