package Int16

import (
	"encoding/binary"
	"strconv"
)

const Len = 2

type Int16 [Len]byte

func New() *Int16 {
	return &Int16{}
}

func (p *Int16) Len() int {
	return Len
}

func (p *Int16) DecodeOne(b []byte) *Int16 {
	p.Decode(b)
	return p
}

func (p *Int16) Decode(b []byte) (out []byte) {
	if len(b) >= Len {
		*p = Int16{b[0], b[1]}
		if len(b) > Len {
			out = b[Len:]
		}
	}
	return
}

func (p *Int16) Encode() []byte {
	return p[:]
}

func (p *Int16) Get() int16 {
	return int16(binary.LittleEndian.Uint16(p[:Len]))
}

func (p *Int16) Put(i int16) *Int16 {
	binary.LittleEndian.PutUint16(p[:Len], uint16(i))
	return p
}

func (p *Int16) String() string {
	return strconv.FormatInt(
		int64(int16(binary.LittleEndian.Uint16(p[:Len]))), 10)
}
