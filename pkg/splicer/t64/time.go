package t64

import (
	"encoding/binary"
	"time"
)

const Len = 8

// Time encodes a standard unix 64 bit, 1-second precision timestamp from a Go
// t64.Time. Note that this omits the monotonic component, it uses the
// time.Unix function to get the "wall clock" time.
type Time struct {
	b []byte
}

// New allocates bytes to store a new Time in. Note that this allocates memory.
func New() *Time { return &Time{b: make([]byte, Len)} }

// NewFrom creates a Time from raw bytes, if the slice is at least Len bytes
// long. This can be used to snip out an encoded segment which should return a
// value from a call to Get.
//
// The remaining bytes, if any, are returned for further processing.
func NewFrom(b []byte) (s *Time, rem []byte) {
	if len(b) < Len {
		return
	}
	// This slices the input, meaning no copy is required, only allocating the
	// slice pointer.
	s = &Time{b: b[:Len]}
	if len(b) > Len {
		rem = b[Len:]
	}
	return
}

func (s *Time) Read() (out []byte) {
	if len(s.b) >= Len {
		out = s.b[:Len]
	}
	return
}

func (s *Time) Write(by []byte) (out []byte) {
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

func (s *Time) Len() int { return len(s.b) }

func (s *Time) Get() (v interface{}) {
	if len(s.b) >= Len {
		t := binary.BigEndian.Uint64(s.b[:Len])
		tv := time.Unix(int64(t), 0)
		v = &tv
	}
	return
}

func (s *Time) Put(t interface{}) interface{} {
	var tv *time.Time
	var ok bool
	if tv, ok = t.(*time.Time); ok {
		binary.BigEndian.PutUint64(s.b[:], uint64(tv.Unix()))
	}
	return s
}

// Assert takes an interface and if it is a t64.Time, returns the time.Time
// value. If it is not the expected type, nil is returned.
func Assert(v interface{}) (t *time.Time) {
	var tv *Time
	var ok bool
	if tv, ok = v.(*Time); ok {
		tt := tv.Get()
		// If this fails the return is nil, indicating failure.
		t, _ = tt.(*time.Time)
	}
	return
}
