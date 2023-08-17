package i32

import "encoding/binary"

// I is a 4 byte value that stores an int32.
type I struct {
	Bytes []byte
}

func New() *I {
	return &I{Bytes: make([]byte, 4)}
}

func (b *I) DecodeOne(by []byte) *I {
	b.Decode(by)
	return b
}

func (b *I) Decode(by []byte) (out []byte) {
	if len(by) >= 4 {
		b.Bytes = []byte{by[0], by[1], by[2], by[3]}
		if len(by) > 4 {
			out = by[4:]
		}
	}
	return
}

func (b *I) Encode() []byte {
	return b.Bytes[:]
}

func (b *I) Get() int32 {
	return int32(binary.BigEndian.Uint32(b.Bytes[:]))
}

func (b *I) Put(bits int32) *I {
	binary.BigEndian.PutUint32(b.Bytes[:], uint32(bits))
	return b
}
