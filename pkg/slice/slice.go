// Package slice is a collection of miscellaneous functions involving slices of
// bytes.
package slice

import (
	"encoding/binary"
)

func Cut(b []byte, l int) (seg []byte, rem []byte) { return b[:l], b[l:] }

func SumLen(chunks ...[]byte) (l int) {
	for _, c := range chunks {
		l += len(c)
	}
	return
}

func Segment(b []byte, segmentSize int) (segs [][]byte) {
	count := len(b) / segmentSize
	if len(b)%segmentSize != 0 {
		count++
	}
	segs = make([][]byte, count)
	for i := 0; i < count; i++ {
		segs[i] = b[i*segmentSize : (i+1)*segmentSize]
	}
	return
}

var put32 = binary.LittleEndian.PutUint32
var get32 = binary.LittleEndian.Uint32
var put16 = binary.LittleEndian.PutUint16
var get16 = binary.LittleEndian.Uint16

// DecodeUint32 returns an int containing the little endian encoded 32bit value
// stored in a 4 byte long slice
func DecodeUint32(b []byte) int { return int(get32(b)) }

// EncodeUint32 puts an int into a uint32 and then into 4 byte long slice.
func EncodeUint32(b []byte, n int) { put32(b, uint32(n)) }

// DecodeUint16 returns an int containing the little endian encoded 32bit value
// stored in a 4 byte long slice
func DecodeUint16(b []byte) int { return int(get16(b)) }

// EncodeUint16 puts an int into a uint32 and then into 2 byte long slice.
func EncodeUint16(b []byte, n int) { put16(b, uint16(n)) }

// Concatenate takes a list of byte slices and packs them together in a packet.
// The returned packet has its capacity pre-allocated to match what gets copied
// into it by append.
func Concatenate(chunks ...[]byte) (pkt []byte) {
	l := SumLen(chunks...)
	pkt = make([]byte, 0, l)
	for i := range chunks {
		pkt = append(pkt, chunks[i]...)
	}
	return
}

const Uint32Len = 4
const Uint16Len = 2

type Size32 []byte
type Size16 []byte

func NewUint32() Size32 { return make(Size32, Uint32Len) }
func NewUint16() Size16 { return make(Size16, Uint16Len) }

func FromHash(b [32]byte) []byte { return b[:] }
