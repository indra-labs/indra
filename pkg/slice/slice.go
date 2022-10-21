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

var PutUint32 = binary.LittleEndian.PutUint32
var Uint32 = binary.LittleEndian.Uint32

// DecodeUint32 returns an int containing the little endian encoded 32bit value
// stored in a 4 byte long slice
func DecodeUint32(b []byte) int    { return int(Uint32(b)) }
func EncodeUint32(b []byte, n int) { PutUint32(b, uint32(n)) }

func Concatenate(chunks ...[]byte) (pkt []byte) {
	for i := range chunks {
		pkt = append(pkt, chunks[i]...)
	}
	return
}

const Len = 4

type Length []byte

func NewLength() Length { return make(Length, 4) }

func FromHash(b [32]byte) []byte { return b[:] }
