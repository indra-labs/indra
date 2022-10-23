package message

import (
	"github.com/Indra-Labs/indra/pkg/slice"
)

// Split creates a series of packets including the defined Reed Solomon
// parameters for extra parity shards.
//
// The last packet that falls short of the segmentSize is padded random bytes.
//
// The segmentSize is inclusive of packet overhead plus the Seen key
// fingerprints at the end of the Packet.
func Split(ep EP, segmentSize int) (pkts [][]byte, e error) {
	overhead := ep.GetOverhead()
	dataSegSize := segmentSize - overhead
	segs := slice.Segment(ep.Data, dataSegSize)
	ls := len(segs)
	pkts = make([][]byte, ls)
	ep.PayloadLen = len(segs[0])
	for i := range segs {
		if i == ls-1 {
			// if this is the last segment it could be truncated, so
			// we want to expand the payload to the same size with
			// noise. The segment parameter will show the real
			// length so the noise will be sliced off by decode.
			ep.PayloadLen = len(segs[i])
			diff := dataSegSize - ep.PayloadLen
			if diff > 0 {
				segs[i] = append(segs[i], slice.NoisePad(diff)...)
			}
		}
		ep.Data = segs[i]
		if pkts[i], e = Encode(ep); check(e) {
			return
		}
	}
	return
}

// Join a cellection of Packet s together.
//
// Every message has a unique sender key, so once a packet is decoded, the
// pub.Print is the key to identifying associated packets.
func Join(pkts []Packet) (msg []byte, e error) {
	return
}
