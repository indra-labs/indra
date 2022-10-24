package message

import (
	"errors"
	"sort"

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
	// In order to do adaptive RS encoding we need to pad out to multiples
	// of 4 to provide 25% increments of redundancy.
	mod := len(segs) % 4
	if mod > 0 {
		extra := make([][]byte, mod)
		segs = append(segs, extra...)
		ls += mod
	}
	pkts = make([][]byte, ls)
	ep.Tot = ls
	for i := range segs {
		ep.Pad = dataSegSize - len(segs[i])
		ep.Data = segs[i]
		ep.Seq = i
		if pkts[i], e = Encode(ep); check(e) {
			return
		}
		ep.Seq++
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
	// determine which, if any, packets are missing, TODO: Reed Solomon FEC.
	var missing []uint16
	prev := pkts[0]
	for _, i := range pkts[1:] {
		if i.Seq-1 != prev.Seq {
			missing = append(missing, i.Seq-1)
		}
		prev = i
	}
	// If none are missing and there is no parity shards we can just zip
	// it all together.
	if len(missing) == 0 && pkts[0].ParityShards == 0 {
		var shards [][]byte
		for i := range pkts {
			shards = append(shards, pkts[i].Payload)
		}
		totalLength := slice.SumLen(shards...)
		msg = make([]byte, 0, totalLength)
		for i := range shards {
			msg = append(msg, shards[i]...)
		}
		return
	}
	e = errors.New("not yet implemented data shard recovery")
	return
}
