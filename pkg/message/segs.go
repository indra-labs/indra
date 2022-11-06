package message

import (
	"fmt"
)

type Segment struct {
	DStart, DEnd, PEnd, SLen, Last int
}

func (s Segment) String() (o string) {
	slast := (s.PEnd-s.DEnd)*s.SLen - s.SLen + s.Last
	if s.PEnd-s.DEnd == 0 {
		slast = 0
	}
	o = fmt.Sprintf("%5d (%5d, %5d) %5d [%5d, %5d] %5d (%5d; %5d)",
		s.DStart, s.DEnd-s.DStart, (s.DEnd-s.DStart-1)*s.SLen+s.Last,
		s.DEnd, s.PEnd-s.DEnd, slast, s.PEnd,
		s.SLen, s.Last)

	return
}

type Segments []Segment

func (s Segments) String() (o string) {
	for _, si := range s {
		o += fmt.Sprintf("\n%s", si.String())
	}
	return
}

func NewSegments(payloadLen, segmentSize, overhead, redundancy int) (s Segments) {
	segSize := segmentSize - overhead
	nSegs := payloadLen/segSize + 1
	lastSeg := payloadLen % segSize
	sectsD := 256 - redundancy
	sectsP := redundancy
	withR := nSegs + nSegs*sectsP/sectsD
	// If any redundancy is specified, if it rounds to zero, it must be
	// bumped up to 1 in order to work with the rs encoder.
	if withR == nSegs && redundancy > 0 {
		withR++
	}
	sects := nSegs / sectsD
	lastSect := nSegs % sectsD
	log.I.F("segSize %d, nSegs %d, lastSeg %d, sectsD %d, sectsP %d, withR %d, sects %d, lastSect %d",
		segSize, nSegs, lastSeg, sectsD, sectsP, withR, sects, lastSect)
	var start int
	for i := 0; i < sects; i++ {
		s = append(s,
			Segment{DStart: start,
				DEnd: start + sectsD,
				PEnd: start + 256,
				SLen: segSize,
				Last: segSize})
		start += 256
	}
	endD := start + lastSect
	// if there is redundancy the DEnd must be at least one less than PEnd.
	if withR == endD && redundancy > 0 {
		withR++
	}
	s = append(s, Segment{
		DStart: start,
		DEnd:   endD,
		PEnd:   withR,
		SLen:   segSize,
		Last:   lastSeg,
	})
	return
}
