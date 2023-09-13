package magic

const Len = 4

type Bytes struct {
	bytes []byte
}

// New allocates bytes to store a new magic.Bytes in. Note that this allocates
// memory.
func New() *Bytes { return &Bytes{bytes: make([]byte, Len)} }

// NewFrom creates a 32-bit integer from raw bytes, if the slice is at least Len
// bytes long. This can be used to snip out an encoded segment which should
// return a value from a call to Get.
//
// The remaining bytes, if any, are returned for further processing.
func NewFrom(b []byte) (s *Bytes, rem []byte) {
	if len(b) < Len {
		return
	}
	// This slices the input, meaning no copy is required, only allocating the
	// slice pointer.
	s = &Bytes{bytes: b[:Len]}
	if len(b) > Len {
		rem = b[Len:]
	}
	return
}

func (m Bytes) Len() (l int) { return len(m.bytes) }

func (m Bytes) Read() (o []byte) {
	if len(m.bytes) >= Len {
		o = m.bytes[:Len]
	}
	return
}

func (m Bytes) Write(b []byte) (o []byte) {
	if len(b) >= Len {
		m.bytes = b[:3]
		// If there is more, return the excess.
		if len(b) > Len {
			o = b[Len:]
		}
	}
	return
}

func (m Bytes) Get() (v interface{}) {
	if len(m.bytes) >= Len {
		val := string(m.bytes[:Len])
		v = &val
	}
	return
}

func (m Bytes) Put(v interface{}) (o interface{}) {
	var tv *string
	var ok bool
	if tv, ok = v.(*string); ok {
		bytes := []byte(*tv)
		if len(bytes) == Len {
			m.bytes = bytes
		}
	}
	return
}

// Assert takes an interface and if it is a duration.Time, returns the time.Time
// value. If it is not the expected type, nil is returned.
func Assert(v interface{}) (t *string) {
	var tv *Bytes
	var ok bool
	if tv, ok = v.(*Bytes); ok {
		tt := tv.Get()
		// If this fails the return is nil, indicating failure.
		t, _ = tt.(*string)
	}
	return
}
