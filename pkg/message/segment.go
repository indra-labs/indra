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
	for i := range segs {
		ep.Pad = dataSegSize - len(segs[i])
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
