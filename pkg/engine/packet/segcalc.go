package packet

import (
	"fmt"
	
	"github.com/templexxx/reedsolomon"
)

type PacketSegment struct {
	DStart, DEnd, PEnd, SLen, Last int
}

// String is a printer that produces a Go syntax formatted version of the
// PacketSegment.
func (s PacketSegment) String() (o string) {
	o = fmt.Sprintf(
		"\t\tSegment{ DStart: %d, DEnd: %d, PEnd: %d, SLen: %d, Last: %d},",
		s.DStart, s.DEnd, s.PEnd, s.SLen, s.Last)
	return
}

type PacketSegments []PacketSegment

// String is a printer that produces a Go syntax formatted version of the
// PacketSegments.
func (s PacketSegments) String() (o string) {
	o += "\n\tSegments{"
	for i := range s {
		o += fmt.Sprintf("\n%s", s[i].String())
	}
	o += "\n\t}\n"
	return
}

func NewPacketSegments(payloadLen, segmentSize, overhead, parity int) (s PacketSegments) {
	segSize := segmentSize - overhead
	nSegs := payloadLen/segSize + 1
	lastSeg := payloadLen % segSize
	sectsD := 256 - parity
	sectsP := parity
	withR := nSegs + nSegs*sectsP/sectsD
	// If any parity is specified, if it rounds to zero, it must be
	// bumped up to 1 in order to work with the rs encoder.
	if withR == nSegs && parity > 0 {
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
				PacketSegment{DStart: start,
					DEnd: start + sectsD,
					PEnd: start + 256,
					SLen: segSize,
					Last: last})
			start += 256
		}
	}
	if lastSect > 0 {
		endD := start + lastSect
		// if there is parity the DEnd must be at least one less
		// than PEnd.
		if withR == endD && parity > 0 {
			withR++
		}
		s = append(s, PacketSegment{
			DStart: start,
			DEnd:   endD,
			PEnd:   withR,
			SLen:   segSize,
			Last:   lastSeg,
		})
	}
	return
}

func (s PacketSegments) AddParity(segs [][]byte) (shards [][]byte, e error) {
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
			if rs, e = reedsolomon.New(dLen, pLen); fails(e) {
				return
			}
			if e = rs.Encode(section); fails(e) {
				return
			}
		}
		shards = append(shards, section...)
	}
	return
}
