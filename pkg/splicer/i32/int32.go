package i32

import "encoding/binary"

const Len = 4

// S is a 4 byte value that stores an int32. S here to signify it is signed.
type S struct {
	b []byte
}

// New allocates bytes to store a new i32.S in. Note that this allocates memory.
func New() *S { return &S{b: make([]byte, Len)} }

// NewFrom creates a 32-bit integer from raw bytes, if the slice is at least Len
// bytes long. This can be used to snip out an encoded segment which should
// return a value from a call to Get.
//
// The remaining bytes, if any, are returned for further processing.
func NewFrom(b []byte) (s *S, rem []byte) {
	// If the bytes are less than Len the input is invalid and nil will be
	// returned.
	if len(b) < Len {
		return
	}
	// This slices the input, meaning no copy is required, only allocating the
	// slice pointer.
	s = &S{b: b[:Len]}
	if len(b) > Len {
		rem = b[Len:]
	}
	return
}

func (s *S) Read() (out []byte) {
	if len(s.b) >= Len {
		out = s.b[:Len]
	}
	return
}

func (s *S) Write(by []byte) (out []byte) {
	if len(by) >= Len {
		s.b = []byte{by[0], by[1], by[2], by[3]}
		if len(by) > Len {
			out = by[Len:]
		}
	}
	return
}

func (s *S) Len() int { return len(s.b) }

func (s *S) Get() (v interface{}) {
	if len(s.b) >= Len {
		val := int32(binary.BigEndian.Uint32(s.b))
		return &val
	}
	return
}

func (s *S) Put(v interface{}) interface{} {
	if len(s.b) < Len {
		s.b = make([]byte, Len)
	}
	var tv *int32
	var ok bool
	if tv, ok = v.(*int32); ok {
		binary.BigEndian.PutUint32(s.b[:Len], uint32(*tv))
	}
	return s
}

// Assert takes an interface and if it is a t64.Time, returns the time.Time
// value. If it is not the expected type, nil is returned.
func Assert(v interface{}) (t *int32) {
	var tv *S
	var ok bool
	if tv, ok = v.(*S); ok {
		tt := tv.Get()
		// If this fails the return is nil, indicating failure.
		t, _ = tt.(*int32)
	}
	return
}
