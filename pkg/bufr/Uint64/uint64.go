package Uint64

import (
	"encoding/binary"
	"strconv"
)

const Len = 8

type Uint64 [Len]byte

func New() *Uint64 {
	return &Uint64{}
}

func (p *Uint64) Len() int {
	return Len
}

func (p *Uint64) DecodeOne(b []byte) *Uint64 {
	p.Decode(b)
	return p
}

func (p *Uint64) Decode(b []byte) (out []byte) {
	if len(b) >= Len {
		*p = Uint64{b[0], b[1], b[2], b[3], b[4], b[5], b[6], b[7]}
		if len(b) > Len {
			out = b[Len:]
		}
	}
	return
}

func (p *Uint64) Encode() []byte {
	return p[:Len]
}

func (p *Uint64) Get() uint64 {
	return binary.LittleEndian.Uint64(p[:Len])
}

func (p *Uint64) Put(i uint64) *Uint64 {
	binary.LittleEndian.PutUint64(p[:Len], i)
	return p
}

func (p *Uint64) String() string {
	return strconv.FormatUint(
		binary.LittleEndian.Uint64(p[:Len]), 10)
}
