package time

import (
	"encoding/binary"
	"time"
)

// Stamp encodes a standard unix 64 bit, 1-second precision timestamp from a Go
// time.Stamp. Note that this omits the monotonic component, it uses the
// time.Unix function to get the "wall clock" time.
type Stamp struct {
	Bytes []byte
}

func New() *Stamp {
	return &Stamp{Bytes: make([]byte, 8)}
}

func (b *Stamp) DecodeOne(by []byte) *Stamp {
	b.Decode(by)
	return b
}

func (b *Stamp) Decode(by []byte) (out []byte) {
	
	if len(by) >= 8 {
		
		out = by[:8]
		b.Bytes = by[:8]
	}
	
	return
}

func (b *Stamp) Encode() []byte {
	return b.Bytes[:]
}

func (b *Stamp) Get() time.Time {
	t := binary.BigEndian.Uint64(b.Bytes[:8])
	return time.Unix(0, int64(t))
}

func (b *Stamp) Put(t time.Time) *Stamp {
	binary.BigEndian.PutUint64(b.Bytes[:], uint64(t.UnixNano()))
	return b
}
