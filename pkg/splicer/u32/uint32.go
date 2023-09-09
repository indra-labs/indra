package u32

import "encoding/binary"

const Len = 4

// U is a 4 byte value that stores an uint32. U here to signify it is unsigned.
type U struct {
	b []byte
}

// New allocates bytes to store a new u32.U in. Note that this allocates memory.
func New() *U { return &U{b: make([]byte, Len)} }

// NewFrom creates a 32-bit integer from raw bytes, if the slice is at least Len
// bytes long. This can be used to snip out an encoded segment which should
// return a value from a call to Get.
//
// The remaining bytes, if any, are returned for further processing.
func NewFrom(b []byte) (s *U, rem []byte) {
	if len(b) < Len {
		return
	}
	// This slices the input, meaning no copy is required, only allocating the
	// slice pointer.
	s = &U{b: b[:Len]}
	if len(b) > Len {
		rem = b[Len:]
	}
	return
}

func (s *U) Read() (out []byte) {
	if len(s.b) >= Len {
		out = s.b[:Len]
	}
	return
}

func (s *U) Write(by []byte) (out []byte) {
	if len(by) >= Len {
		s.b = []byte{by[0], by[1], by[2], by[3]}
		if len(by) > Len {
			out = by[Len:]
		}
	}
	return
}

func (s *U) Len() int { return len(s.b) }

func (s *U) Get() (v interface{}) {
	val := binary.BigEndian.Uint32(s.b)
	return &val
}

func (s *U) Put(bits interface{}) interface{} {
	var tv *uint32
	var ok bool
	if tv, ok = bits.(*uint32); ok {
		binary.BigEndian.PutUint32(s.b[:Len], *tv)
	}
	return s
}

// Assert takes an interface and if it is a duration.Time, returns the time.Time
// value. If it is not the expected type, nil is returned.
func Assert(v interface{}) (t *uint32) {
	var tv *U
	var ok bool
	if tv, ok = v.(*U); ok {
		tt := tv.Get()
		// If this fails the return is nil, indicating failure.
		t, _ = tt.(*uint32)
	}
	return
}
