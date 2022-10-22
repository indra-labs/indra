package message

import (
	"github.com/Indra-Labs/indra/pkg/key/pub"
	"github.com/Indra-Labs/indra/pkg/slice"
)

// Split creates a series of packets including the defined Reed Solomon
// parameters for extra parity shards.
//
// The last packet that falls short of the segmentSize is padded with
// deterministic noise to make it fit.
//
// The segmentSize is inclusive of packet overhead, PacketDataMinSize, plus the
// Seen key fingerprints at the end of the Packet.
func Split(ep EP, segmentSize int) (pkts [][]byte, e error) {
	data := ep.Data
	overhead := PacketDataMinSize + len(ep.Seen)*pub.PrintLen
	dataSegSize := segmentSize - overhead
	segs := slice.Segment(data, dataSegSize)
	pkts = make([][]byte, len(segs))
	for i := range segs {
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
