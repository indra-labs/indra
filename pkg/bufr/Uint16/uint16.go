package Uint16

import (
	"encoding/binary"
	"strconv"
)

const Len = 2

type Uint16 [Len]byte

func New() *Uint16 {
	return &Uint16{}
}

func (p *Uint16) Len() int {
	return Len
}

func (p *Uint16) DecodeOne(b []byte) *Uint16 {
	p.Decode(b)
	return p
}

func (p *Uint16) Decode(b []byte) (out []byte) {
	if len(b) >= Len {
		*p = Uint16{b[0], b[1]}
		if len(b) > Len {
			out = b[Len:]
		}
	}
	return
}

func (p *Uint16) Encode() []byte {
	return p[:]
}

func (p *Uint16) Get() uint16 {
	return binary.LittleEndian.Uint16(p[:Len])
}

func (p *Uint16) Put(i uint16) *Uint16 {
	binary.LittleEndian.PutUint16(p[:Len], i)
	return p
}

func (p *Uint16) String() string {
	return strconv.FormatUint(
		uint64(binary.LittleEndian.Uint16(p[:Len])), 10)
}
