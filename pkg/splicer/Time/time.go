package Time

import (
	"encoding/binary"
	"time"
)

// Time encodes a standard unix 64 bit, 1-second precision timestamp from a Go
// time.Time. Note that this omits the monotonic component, it uses the
// time.Unix function to get the "wall clock" time.
type Time struct {
	Bytes [8]byte
}

func New() *Time {
	return &Time{}
}

func (b *Time) DecodeOne(by []byte) *Time {
	b.Decode(by)
	return b
}

func (b *Time) Decode(by []byte) (out []byte) {
	if len(by) >= 4 {
		b.Bytes = [8]byte{by[0], by[1], by[2], by[3], by[4], by[5], by[6],
			by[7]}
		if len(by) > 8 {
			out = by[:8]
		}
	}
	return
}

func (b *Time) Encode() []byte {
	return b.Bytes[:]
}

func (b *Time) Get() time.Time {
	t := binary.BigEndian.Uint64(b.Bytes[:8])
	return time.Unix(0, int64(t))
}

func (b *Time) Put(t time.Time) *Time {
	binary.BigEndian.PutUint64(b.Bytes[:], uint64(t.UnixNano()))
	return b
}
