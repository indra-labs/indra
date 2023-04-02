package engine

import (
	"fmt"
	"testing"
)

var Expected = []string{
	`
	SpliceSegments{
		PacketSegment{ DStart: 0, DEnd: 192, PEnd: 256, SLen: 195, ID: 195},
		PacketSegment{ DStart: 256, DEnd: 448, PEnd: 512, SLen: 195, ID: 195},
		PacketSegment{ DStart: 512, DEnd: 704, PEnd: 768, SLen: 195, ID: 195},
		PacketSegment{ DStart: 768, DEnd: 960, PEnd: 1024, SLen: 195, ID: 195},
		PacketSegment{ DStart: 1024, DEnd: 1216, PEnd: 1280, SLen: 195, ID: 195},
		PacketSegment{ DStart: 1280, DEnd: 1472, PEnd: 1536, SLen: 195, ID: 195},
		PacketSegment{ DStart: 1536, DEnd: 1728, PEnd: 1792, SLen: 195, ID: 195},
		PacketSegment{ DStart: 1792, DEnd: 1793, PEnd: 1794, SLen: 195, ID: 175},
	}
`,
	`
	SpliceSegments{
		PacketSegment{ DStart: 0, DEnd: 130, PEnd: 130, SLen: 4035, ID: 3773},
	}
`,
	`
	SpliceSegments{
		PacketSegment{ DStart: 0, DEnd: 128, PEnd: 256, SLen: 4035, ID: 4035},
		PacketSegment{ DStart: 256, DEnd: 258, PEnd: 260, SLen: 4035, ID: 3773},
	}
`,
	`
	SpliceSegments{
		PacketSegment{ DStart: 0, DEnd: 65, PEnd: 65, SLen: 4035, ID: 3904},
	}
`,
}

func TestNewSegments(t *testing.T) {
	msgSize := 2<<17 + 111
	segSize := 256
	s := NewPacketSegments(msgSize, segSize, PacketOverhead, 64)
	o := fmt.Sprint(s)
	if o != Expected[0] {
		t.Errorf(
			"Failed to correctly generate.\ngot:\n'%s'\nexpected:\n'%s'",
			o, Expected[0])
	}
	msgSize = 2 << 18
	segSize = 4096
	s = NewPacketSegments(msgSize, segSize, PacketOverhead, 0)
	o = fmt.Sprint(s)
	if o != Expected[1] {
		t.Errorf(
			"Failed to correctly generate.\ngot:\n%s\nexpected:\n%s",
			o, Expected[1])
	}
	s = NewPacketSegments(msgSize, segSize, PacketOverhead, 128)
	o = fmt.Sprint(s)
	if o != Expected[2] {
		t.Errorf(
			"Failed to correctly generate.\ngot:\n%s\nexpected:\n%s",
			o, Expected[2])
	}
	msgSize = 2 << 17
	segSize = 4096
	s = NewPacketSegments(msgSize, segSize, PacketOverhead, 0)
	o = fmt.Sprint(s)
	if o != Expected[3] {
		t.Errorf(
			"Failed to correctly generate.\ngot:\n%s\nexpected:\n%s",
			o, Expected[3])
	}
}
