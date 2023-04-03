package packet

import (
	"fmt"
	
	"github.com/templexxx/reedsolomon"
)

type Segment struct {
	DStart, DEnd, PEnd, SLen, Last int
}

// This is an expanded printer for debugging
// func (s Segment) String() (o string) {
// 	slast := (s.PEnd-s.DEnd)*s.SLen - s.SLen + s.ID
// 	if s.PEnd-s.DEnd == 0 {
// 		slast = 0
// 	}
// 	o = fmt.Sprintf("%5d (%5d, %5d) %5d [%5d, %5d] %5d (%5d; %5d)",
// 		s.DStart, s.DEnd-s.DStart, (s.DEnd-s.DStart-1)*s.SLen+s.ID,
// 		s.DEnd, s.PEnd-s.DEnd, slast, s.PEnd,
// 		s.SLen, s.ID)
// 	return
// }

// String is a printer that produces a Go syntax formatted version of the
// Segment.
func (s Segment) String() (o string) {
	o = fmt.Sprintf(
		"\t\tSegment{ DStart: %d, DEnd: %d, PEnd: %d, SLen: %d, ID: %d},", s.DStart, s.DEnd, s.PEnd, s.SLen, s.Last)
	return
}

type Segments []Segment

// String is a printer that produces a Go syntax formatted version of the
// Segments.
func (s Segments) String() (o string) {
	o += "\n\tSegments{"
	for _, si := range s {
		o += fmt.Sprintf("\n%s", si.String())
	}
	o += "\n\t}\n"
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
	var start int
	if sects > 0 {
		last := segSize
		if lastSect == 0 && sects == 1 {
			last = lastSeg
		}
		for i := 0; i < sects; i++ {
			s = append(s,
				Segment{DStart: start,
					DEnd: start + sectsD,
					PEnd: start + 256,
					SLen: segSize,
					Last: last})
			start += 256
		}
	}
	if lastSect > 0 {
		endD := start + lastSect
		// if there is redundancy the DEnd must be at least one less
		// than PEnd.
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
	}
	return
}

func (s Segments) AddParity(segs [][]byte) (shards [][]byte, e error) {
	var segLen int
	for i := range s {
		segLen += s[i].DEnd - s[i].DStart
	}
	if len(segs) != segLen {
		e = fmt.Errorf("slice wrong length, got %d expected %d",
			len(segs), segLen)
	}
	for i := range s {
		dLen := s[i].DEnd - s[i].DStart
		pLen := s[i].PEnd - s[i].DEnd
		section := make([][]byte, 0, dLen+pLen)
		section, segs = append(section, segs[:dLen]...), segs[dLen:]
		for j := 0; j < pLen; j++ {
			section = append(section, make([]byte, s[i].SLen))
		}
		if pLen > 0 {
			var rs *reedsolomon.RS
			if rs, e = reedsolomon.New(dLen, pLen); check(e) {
				return
			}
			if e = rs.Encode(section); check(e) {
				return
			}
		}
		shards = append(shards, section...)
	}
	return
}
