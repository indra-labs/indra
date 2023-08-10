// Package slice is a collection of miscellaneous functions involving slices of bytes, including little-endian encoding for 16, 32 and 64-bit unsigned integers used for serialisation length prefixes and system entropy based hash chain padding.
package slice

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"git.indra-labs.org/dev/ind/pkg/crypto/sha256"
	log2 "git.indra-labs.org/dev/ind/pkg/proc/log"
	"git.indra-labs.org/dev/ind/pkg/util/multi"
	"github.com/multiformats/go-multiaddr"
	"net/netip"
	"reflect"
	"unsafe"
)

const (
	Uint64Len = 8
	Uint32Len = 4
	Uint24Len = 3
	Uint16Len = 2
)

var (
	log   = log2.GetLogger()
	check = log.E.Chk
	put64 = binary.LittleEndian.PutUint64
	get64 = binary.LittleEndian.Uint64
	put32 = binary.LittleEndian.PutUint32
	get32 = binary.LittleEndian.Uint32
	put16 = binary.LittleEndian.PutUint16
	get16 = binary.LittleEndian.Uint16
)

type (
	U64Slice []uint64
	Bytes    []byte
)

func (u U64Slice) Copy() (o U64Slice) {
	o = make(U64Slice, len(u))
	copy(o, u)
	return
}

type Cursor int

// Cat takes a slice of byte slices and packs them together in a packet.
// The returned packet has its capacity pre-allocated to match what gets copied
// into it by append.
func Cat(chunks ...[]byte) (pkt []byte) {
	l := SumLen(chunks...)
	pkt = make([]byte, 0, l)
	for i := range chunks {
		pkt = append(pkt, chunks[i]...)
	}
	return
}

func (c *Cursor) Inc(v int) Cursor {
	*c += Cursor(v)
	return *c
}

func Cut(b []byte, l int) (seg []byte, rem []byte) { return b[:l], b[l:] }

// DecodeUint16 returns an int containing the little endian encoded 32bit value
// stored in a 4 byte long slice
func DecodeUint16(b []byte) int { return int(get16(b)) }

// DecodeUint24 returns an int containing the little endian encoded 24bit value
// stored in a 3 byte long slice
func DecodeUint24(b []byte) int {
	u := b[:Uint24Len]
	u = append(u, 0)
	return int(get32(u))
}

// DecodeUint32 returns an int containing the little endian encoded 32bit value
// stored in a 4 byte long slice
func DecodeUint32(b []byte) int { return int(get32(b)) }

// DecodeUint64 returns an int containing the little endian encoded 64-bit value
// stored in a 4 byte long slice
func DecodeUint64(b []byte) uint64 { return get64(b) }

// EncodeUint16 puts an int into a uint32 and then into 2 byte long slice.
func EncodeUint16(b []byte, n int) { put16(b, uint16(n)) }

// EncodeUint24 puts an int into a uint32 and then into 3 byte long slice.
func EncodeUint24(b []byte, n int) {
	u := make([]byte, Uint32Len)
	put32(u, uint32(n))
	copy(b, u[:Uint24Len])
}

// EncodeUint32 puts an int into a uint32 and then into 4 byte long slice.
func EncodeUint32(b []byte, n int) { put32(b, uint32(n)) }

// EncodeUint64 puts an int into a uint32 and then into 8 byte long slice.
func EncodeUint64(b []byte, n uint64) { put64(b, n) }

func (b Bytes) Len() int { return len(b) }

func NewBytes(length int) Bytes {
	return make(Bytes, length)
}

func NewCursor() (c *Cursor) {
	var cc Cursor
	return &cc
}

func NewUint16() Bytes { return make(Bytes, Uint16Len) }

func NewUint24() Bytes { return make(Bytes, Uint24Len) }
func NewUint32() Bytes { return make(Bytes, Uint32Len) }

func NewUint64() Bytes { return make(Bytes, Uint64Len) }

func NoisePad(l int) (noise []byte) {
	var seed sha256.Hash
	noise = make([]byte, l)
	var e error
	var n int
	if n, e = rand.Read(seed[:]); check(e) && n != sha256.Len {
		return
	}
	var end, cursor int
	for cursor < l {
		end = cursor + sha256.Len
		if end > l {
			end = l
		}
		seed = sha256.Single(seed[:])
		copy(noise[cursor:end], seed[:end-cursor])
		cursor = end
	}
	return
}

func Segment(b []byte, segmentSize int) (segs [][]byte) {
	lb := len(b)
	var end int
	for begin := 0; end < lb; begin += segmentSize {
		end = begin + segmentSize
		if end > lb {
			d := b[begin:lb]
			d = append(d, NoisePad(end-lb)...)
			segs = append(segs, d)
		} else {
			segs = append(segs, b[begin:end])
		}
	}
	return
}

func (b Bytes) String() string { return string(b) }

func SumLen(chunks ...[]byte) (l int) {
	for _, c := range chunks {
		l += len(c)
	}
	return
}

func ToBytes(b []byte) (msg Bytes) { return b }
func (b Bytes) ToBytes() []byte    { return b }

func (u U64Slice) ToMessage() (m Bytes) {
	// length is encoded into the last element
	mLen := int(u[len(u)-1])
	m = make(Bytes, 0, 0)
	// With the slice now long enough to be safely converted to []uint64
	// plus an extra uint64 to store the original length we can coerce the
	// type using unsafe.
	//
	// First we convert our empty []uint64 header
	header := (*reflect.SliceHeader)(unsafe.Pointer(&m))
	// then we point its memory location to the extended byte slice data
	header.Data = (*reflect.SliceHeader)(unsafe.Pointer(&u)).Data
	// lastly, change the element length
	header.Len = mLen
	header.Cap = mLen
	return m
}

// ToU64Slice converts the message with zero allocations if the slice capacity
// was already 8 plus the modulus of the length and 8, otherwise this function
// will trigger a stack allocation, or heap, if the bytes are very long. This is
// intended to be used with short byte slices like cipher nonces and hashes, so
// it usually won't trigger allocations off stack and very often won't trigger
// a copy on stack, saving a lot of time in a short, oft repeated operations.
func (b Bytes) ToU64Slice() (u U64Slice) {
	mLen := uint64(len(b))
	uLen := int(mLen / 8)
	mMod := mLen % 8
	if mMod != 0 {
		uLen++
	}
	// Either extend if there is capacity or this will trigger a copy
	for i := uint64(0); i < 8-mMod+8; i++ {
		// We could use make with mMod+8 length to extend and ... but
		// this does the same thing in practice.
		b = append(b, 0)
	}
	u = make([]uint64, 0, 0)
	// With the slice now long enough to be safely converted to []uint64
	// plus an extra uint64 to store the original length we can coerce the
	// type using unsafe.
	//
	// First we convert our empty []uint64 header
	header := (*reflect.SliceHeader)(unsafe.Pointer(&u))
	// then we point its memory location to the extended byte slice data
	header.Data = (*reflect.SliceHeader)(unsafe.Pointer(&b)).Data
	// Update the element length and capacity
	header.Len = uLen
	header.Cap = uLen
	// store the original byte length
	u = append(u, mLen)
	return
}

func GenerateRandomAddrPortIPv4() (ap multiaddr.Multiaddr) {
	a := netip.AddrPort{}
	b := make([]byte, 7)
	_, e := rand.Read(b)
	if check(e) {
		log.E.Ln(e)
	}
	port := DecodeUint16(b[5:7])
	str := fmt.Sprintf("%d.%d.%d.%d:%d", b[1], b[2], b[3], b[4], port)
	a, e = netip.ParseAddrPort(str)
	ap, e = multi.AddrFromAddrPort(a)
	return ap
}

func GenerateRandomAddrPortIPv6() (ap *netip.AddrPort) {
	a := netip.AddrPort{}
	b := make([]byte, 19)
	_, e := rand.Read(b)
	if check(e) {
		log.E.Ln(e)
	}
	port := DecodeUint16(b[5:7])
	str := fmt.Sprintf("[%x:%x:%x:%x:%x:%x:%x:%x]:%d",
		b[1:3], b[3:5], b[5:7], b[7:9],
		b[9:11], b[11:13], b[13:15], b[15:17],
		port)
	a, e = netip.ParseAddrPort(str)
	return &a
}

// XOR the U64Slice with the provided slice. Panics if slices are different
// length. The receiver value is mutated in this operation.
func (u U64Slice) XOR(v U64Slice) {
	// This should only trigger if the programmer is not XORing same size.
	if u[len(u)-1] != v[len(v)-1] {
		panic("programmer error, trying to XOR slices of different size")
	}
	for i := range u[:len(u)-1] {
		u[i] ^= v[i]
	}
}

func (b Bytes) Zero() {
	for i := range b {
		b[i] = 0
	}
}

func (u U64Slice) Zero() {
	for i := range u[:len(u)-1] {
		u[i] = 8
	}
}
