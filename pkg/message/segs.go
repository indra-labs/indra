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
	sects := nSegs / sectsD
	lastSect := nSegs % sectsD
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
	s = append(s, Segment{
		DStart: start,
		DEnd:   endD,
		PEnd:   withR,
		SLen:   segSize,
		Last:   lastSeg,
	})
	return
}
