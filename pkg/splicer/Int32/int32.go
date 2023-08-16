package Int32

import "encoding/binary"

// Int32 is a 32 bit value that stores an int32 (used for block height).
// I don't think the sign is preserved but block heights are never negative
// except with special semantics
type Int32 struct {
	Bytes [4]byte
}

func New() *Int32 {
	return &Int32{}
}

func (b *Int32) DecodeOne(by []byte) *Int32 {
	b.Decode(by)
	return b
}

func (b *Int32) Decode(by []byte) (out []byte) {
	if len(by) >= 4 {
		b.Bytes = [4]byte{by[0], by[1], by[2], by[3]}
		if len(by) > 4 {
			out = by[4:]
		}
	}
	return
}

func (b *Int32) Encode() []byte {
	return b.Bytes[:]
}

func (b *Int32) Get() int32 {
	return int32(binary.BigEndian.Uint32(b.Bytes[:]))
}

func (b *Int32) Put(bits int32) *Int32 {
	binary.BigEndian.PutUint32(b.Bytes[:], uint32(bits))
	return b
}
