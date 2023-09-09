package duration

import (
	"encoding/binary"
	"time"
)

const Len = 8

// T encodes a duration in nanoseconds as a 64 bit, 8 byte value.
type T struct {
	b []byte
}

// New allocates bytes to store a new T in. Note that this allocates memory.
func New() *T { return &T{b: make([]byte, Len)} }

// NewFrom creates a T from raw bytes, if the slice is at least Len bytes
// long. This can be used to snip out an encoded segment which should return a
// value from a call to Get.
//
// The remaining bytes, if any, are returned for further processing.
func NewFrom(b []byte) (s *T, rem []byte) {
	if len(b) < Len {
		return
	}
	// This slices the input, meaning no copy is required, only allocating the
	// slice pointer.
	s = &T{b: b[:Len]}
	if len(b) > Len {
		rem = b[Len:]
	}
	return
}

func (s *T) Read() (out []byte) {
	if len(s.b) >= Len {
		out = s.b[:Len]
	}
	return
}

func (s *T) Write(by []byte) (out []byte) {
	if len(by) >= Len {
		// If there is already allocated bytes, copy instead of assign.
		if len(s.b) >= Len {
			copy(s.b[:Len], by[:Len])
		} else {
			s.b = by[:Len]
		}
		out = by[Len:]
	}
	return
}

func (s *T) Len() int { return len(s.b) }

func (s *T) Get() (v interface{}) {
	if len(s.b) >= Len {
		t := binary.BigEndian.Uint64(s.b[:Len])
		tv := time.Duration(int64(t))
		v = &tv
	}
	return
}

func (s *T) Put(t interface{}) interface{} {
	var tv *time.Duration
	var ok bool
	if tv, ok = t.(*time.Duration); ok {
		binary.BigEndian.PutUint64(s.b[:], uint64(*tv))
	}
	return s
}

// Assert takes an interface and if it is a duration.T, returns the time.T
// value. If it is not the expected type, nil is returned.
func Assert(v interface{}) (t *time.Duration) {
	var tv *T
	var ok bool
	if tv, ok = v.(*T); ok {
		tt := tv.Get()
		// If this fails the return is nil, indicating failure.
		t, _ = tt.(*time.Duration)
	}
	return
}
