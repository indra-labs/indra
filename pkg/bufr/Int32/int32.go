package Int32

import (
	"encoding/binary"
	"strconv"
)

const Len = 4

type Int32 [Len]byte

func New() *Int32 {
	return &Int32{}
}

func (p *Int32) Len() int {
	return Len
}

func (p *Int32) DecodeOne(b []byte) *Int32 {
	p.Decode(b)
	return p
}

func (p *Int32) Decode(b []byte) (out []byte) {
	if len(b) >= Len {
		*p = Int32{b[0], b[1], b[2], b[3]}
		if len(b) > Len {
			out = b[Len:]
		}
	}
	return
}

func (p *Int32) Encode() []byte {
	return p[:Len]
}

func (p *Int32) Get() int32 {
	return int32(binary.LittleEndian.Uint32(p[:Len]))
}

func (p *Int32) Put(i int32) *Int32 {
	binary.LittleEndian.PutUint32(p[:Len], uint32(i))
	return p
}

func (p *Int32) String() string {
	return strconv.FormatInt(
		int64(int32(binary.LittleEndian.Uint32(p[:Len]))), 10)
}
