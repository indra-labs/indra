package message

import (
	"testing"
)

func TestNewSegments(t *testing.T) {
	msgSize := 2<<17 + 111
	segSize := 256
	s := NewSegments(msgSize, segSize, Overhead, 64)
	log.I.Ln(s)
	msgSize = 2 << 18
	segSize = 4096
	s = NewSegments(msgSize, segSize, Overhead, 0)
	log.I.Ln(s)
	s = NewSegments(msgSize, segSize, Overhead, 128)
	log.I.Ln(s)
	msgSize = 2 << 17
	segSize = 4096
	s = NewSegments(msgSize, segSize, Overhead, 0)
	log.I.Ln(s)
}
