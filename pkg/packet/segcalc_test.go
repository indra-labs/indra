package packet

import (
	"fmt"
	"testing"
)

var Expected = []string{
	`
	Segments{
		Segment{ DStart: 0, DEnd: 192, PEnd: 256, SLen: 203, Last: 203},
		Segment{ DStart: 256, DEnd: 448, PEnd: 512, SLen: 203, Last: 203},
		Segment{ DStart: 512, DEnd: 704, PEnd: 768, SLen: 203, Last: 203},
		Segment{ DStart: 768, DEnd: 960, PEnd: 1024, SLen: 203, Last: 203},
		Segment{ DStart: 1024, DEnd: 1216, PEnd: 1280, SLen: 203, Last: 203},
		Segment{ DStart: 1280, DEnd: 1472, PEnd: 1536, SLen: 203, Last: 203},
		Segment{ DStart: 1536, DEnd: 1676, PEnd: 1722, SLen: 203, Last: 182},
	}
`,
	`
	Segments{
		Segment{ DStart: 0, DEnd: 130, PEnd: 130, SLen: 4043, Last: 2741},
	}
`,
	`
	Segments{
		Segment{ DStart: 0, DEnd: 128, PEnd: 256, SLen: 4043, Last: 4043},
		Segment{ DStart: 256, DEnd: 258, PEnd: 260, SLen: 4043, Last: 2741},
	}
`,
	`
	Segments{
		Segment{ DStart: 0, DEnd: 65, PEnd: 65, SLen: 4043, Last: 3392},
	}
`,
}

func TestNewSegments(t *testing.T) {
	msgSize := 2<<17 + 111
	segSize := 256
	s := NewSegments(msgSize, segSize, Overhead, 64)
	o := fmt.Sprint(s)
	if o != Expected[0] {
		t.Errorf(
			"Failed to correctly generate.\ngot:\n'%s'\nexpected:\n'%s'",
			o, Expected[0])
	}
	msgSize = 2 << 18
	segSize = 4096
	s = NewSegments(msgSize, segSize, Overhead, 0)
	o = fmt.Sprint(s)
	if o != Expected[1] {
		t.Errorf(
			"Failed to correctly generate.\ngot:\n%s\nexpected:\n%s",
			o, Expected[1])
	}
	s = NewSegments(msgSize, segSize, Overhead, 128)
	o = fmt.Sprint(s)
	if o != Expected[2] {
		t.Errorf(
			"Failed to correctly generate.\ngot:\n%s\nexpected:\n%s",
			o, Expected[2])
	}
	msgSize = 2 << 17
	segSize = 4096
	s = NewSegments(msgSize, segSize, Overhead, 0)
	o = fmt.Sprint(s)
	if o != Expected[3] {
		t.Errorf(
			"Failed to correctly generate.\ngot:\n%s\nexpected:\n%s",
			o, Expected[3])
	}
}
