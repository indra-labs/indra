package segment

import (
	"fmt"
	"testing"

	"github.com/Indra-Labs/indra/pkg/packet"
)

var Expected = []string{
	`
	Segments{
		Segment{ DStart: 0, DEnd: 192, PEnd: 256, SLen: 156, Last: 156},
		Segment{ DStart: 256, DEnd: 448, PEnd: 512, SLen: 156, Last: 156},
		Segment{ DStart: 512, DEnd: 704, PEnd: 768, SLen: 156, Last: 156},
		Segment{ DStart: 768, DEnd: 960, PEnd: 1024, SLen: 156, Last: 156},
		Segment{ DStart: 1024, DEnd: 1216, PEnd: 1280, SLen: 156, Last: 156},
		Segment{ DStart: 1280, DEnd: 1472, PEnd: 1536, SLen: 156, Last: 156},
		Segment{ DStart: 1536, DEnd: 1728, PEnd: 1792, SLen: 156, Last: 156},
		Segment{ DStart: 1792, DEnd: 1984, PEnd: 2048, SLen: 156, Last: 156},
		Segment{ DStart: 2048, DEnd: 2194, PEnd: 2242, SLen: 156, Last: 19},
	}
`,
	`
	Segments{
		Segment{ DStart: 0, DEnd: 132, PEnd: 132, SLen: 3996, Last: 812},
	}
`,
	`
	Segments{
		Segment{ DStart: 0, DEnd: 128, PEnd: 256, SLen: 3996, Last: 3996},
		Segment{ DStart: 256, DEnd: 260, PEnd: 264, SLen: 3996, Last: 812},
	}
`,
	`
	Segments{
		Segment{ DStart: 0, DEnd: 66, PEnd: 66, SLen: 3996, Last: 2404},
	}
`,
}

func TestNewSegments(t *testing.T) {
	msgSize := 2<<17 + 111
	segSize := 256
	s := NewSegments(msgSize, segSize, packet.Overhead, 64)
	o := fmt.Sprint(s)
	if o != Expected[0] {
		t.Errorf(
			"Failed to correctly generate.\ngot:\n'%s'\nexpected:\n'%s'",
			o, Expected[0])
	}
	msgSize = 2 << 18
	segSize = 4096
	s = NewSegments(msgSize, segSize, packet.Overhead, 0)
	o = fmt.Sprint(s)
	if o != Expected[1] {
		t.Errorf(
			"Failed to correctly generate.\ngot:\n%s\nexpected:\n%s",
			o, Expected[0])
	}
	s = NewSegments(msgSize, segSize, packet.Overhead, 128)
	o = fmt.Sprint(s)
	if o != Expected[2] {
		t.Errorf(
			"Failed to correctly generate.\ngot:\n%s\nexpected:\n%s",
			o, Expected[0])
	}
	msgSize = 2 << 17
	segSize = 4096
	s = NewSegments(msgSize, segSize, packet.Overhead, 0)
	o = fmt.Sprint(s)
	if o != Expected[3] {
		t.Errorf(
			"Failed to correctly generate.\ngot:\n%s\nexpected:\n%s",
			o, Expected[0])
	}
}
