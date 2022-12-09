package ifc

import (
	"reflect"
	"unsafe"

	"github.com/Indra-Labs/indra"
	log2 "github.com/cybriq/proc/pkg/log"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

type Transport interface {
	Send(b Message)
	Receive() <-chan Message
}

type Message []byte

func ToMessage(b []byte) (msg Message) { return b }
func (m Message) ToBytes() []byte      { return m }

type U64Slice []uint64

// ToU64Slice converts the message with zero allocations if the slice capacity
// was already 8 plus the modulus of the length and 8, otherwise this function
// will trigger a stack allocation, or heap, if the bytes are very long. This is
// intended to be used with short byte slices like cipher nonces and hashes, so
// it usually won't trigger allocations off stack and very often won't trigger
// a copy on stack, saving a lot of time in a short, oft repeated operations.
func (m Message) ToU64Slice() (u U64Slice) {
	mLen := uint64(len(m))
	uLen := int(mLen / 8)
	mMod := mLen % 8
	if mMod != 0 {
		uLen++
	}
	// Either extend if there is capacity or this will trigger a copy
	for i := uint64(0); i < 8-mMod+8; i++ {
		// We could use make with mMod+8 length to extend and ... but
		// this does the same thing in practice.
		m = append(m, 0)
	}
	u = make([]uint64, 0, 0)
	// With the slice now long enough to be safely converted to []uint64
	// plus an extra uint64 to store the original length we can coerce the
	// type using unsafe.
	//
	// First we convert our empty []uint64 header
	header := (*reflect.SliceHeader)(unsafe.Pointer(&u))
	// then we point its memory location to the extended byte slice data
	header.Data = (*reflect.SliceHeader)(unsafe.Pointer(&m)).Data
	// lastly, change the element length
	header.Len = uLen
	// vh := reflect.ValueOf(header)
	// vh.SetLen(len(m) / 8)
	// This safely pins the total capacity to the length, regardless of what
	// it was before. Simply setting the value could cause the GC confusion.
	// vh.SetCap(header.Len)
	header.Cap = uLen
	// store the original byte length
	u = append(u, mLen)
	return
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

func (u U64Slice) ToMessage() (m Message) {
	// log.I.S(u)
	// length is encoded into the last element
	mLen := int(u[len(u)-1])
	m = make(Message, 0, 0)
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
	// vh := reflect.ValueOf(header)
	// vh.SetLen(len(m) / 8)
	// This safely pins the total capacity to the length, regardless of what
	// it was before. Simply setting the value could cause the GC confusion.
	// vh.SetCap(header.Len)
	header.Cap = mLen
	return m
}
