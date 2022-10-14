package Uint32

import (
	"encoding/binary"
	"strconv"
)

const Len = 4

type Uint32 [Len]byte

func New() *Uint32 {
	return &Uint32{}
}

func (p *Uint32) Len() int {
	return Len
}

func (p *Uint32) DecodeOne(b []byte) *Uint32 {
	p.Decode(b)
	return p
}

func (p *Uint32) Decode(b []byte) (out []byte) {
	if len(b) >= Len {
		*p = Uint32{b[0], b[1], b[2], b[3]}
		if len(b) > Len {
			out = b[Len:]
		}
	}
	return
}

func (p *Uint32) Encode() []byte {
	return p[:Len]
}

func (p *Uint32) Get() uint32 {
	return binary.LittleEndian.Uint32(p[:Len])
}

func (p *Uint32) Put(i uint32) *Uint32 {
	binary.LittleEndian.PutUint32(p[:Len], i)
	return p
}

func (p *Uint32) String() string {
	return strconv.FormatUint(
		uint64(binary.LittleEndian.Uint32(p[:Len])), 10)
}
