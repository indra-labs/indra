package message

import (
	"errors"
	"fmt"
	"sort"

	"github.com/Indra-Labs/indra/pkg/slice"
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
	ep.Length = len(ep.Data)
	overhead := ep.GetOverhead()
	ss := segSize - overhead
	segments := slice.Segment(ep.Data, ss)
	segMap := NewSegments(ep.Length, segSize, ep.GetOverhead(), ep.Parity)
	var p [][]byte
	p, e = segMap.AddParity(segments)
	for i := range p {
		ep.Data, ep.Seq = p[i], i
		var s []byte
		if s, e = Encode(ep); check(e) {
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
	_ = lp
	return
}

func RemovePacket(slice Packets, s int) Packets {
	return append(slice[:s], slice[s+1:]...)
}
