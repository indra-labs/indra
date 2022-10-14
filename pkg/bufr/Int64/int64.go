package Uint64

import (
	"encoding/binary"
	"strconv"
)

const Len = 8

type Int64 [Len]byte

func New() *Int64 {
	return &Int64{}
}

func (p *Int64) Len() int {
	return Len
}

func (p *Int64) DecodeOne(b []byte) *Int64 {
	p.Decode(b)
	return p
}

func (p *Int64) Decode(b []byte) (out []byte) {
	if len(b) >= Len {
		*p = Int64{b[0], b[1], b[2], b[3], b[4], b[5], b[6], b[7]}
		if len(b) > Len {
			out = b[Len:]
		}
	}
	return
}

func (p *Int64) Encode() []byte {
	return p[:Len]
}

func (p *Int64) Get() int64 {
	return int64(binary.LittleEndian.Uint64(p[:Len]))
}

func (p *Int64) Put(i int64) *Int64 {
	binary.LittleEndian.PutUint64(p[:Len], uint64(i))
	return p
}

func (p *Int64) String() string {
	return strconv.FormatInt(
		int64(binary.LittleEndian.Uint64(p[:Len])), 10)
}
