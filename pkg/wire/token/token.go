package token

import (
	"github.com/Indra-Labs/indra"
	"github.com/Indra-Labs/indra/pkg/sha256"
	"github.com/Indra-Labs/indra/pkg/slice"
	"github.com/Indra-Labs/indra/pkg/types"
	"github.com/Indra-Labs/indra/pkg/wire/magicbytes"
	log2 "github.com/cybriq/proc/pkg/log"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

// A Type is a 32 byte value.
type Type sha256.Hash

var (
	Magic              = slice.Bytes("tok")
	MinLen             = magicbytes.Len + sha256.Len
	_      types.Onion = Type{}
)

func (x Type) Inner() types.Onion   { return nil }
func (x Type) Insert(_ types.Onion) {}
func (x Type) Len() int             { return MinLen }

func (x Type) Encode(o slice.Bytes, c *slice.Cursor) {
	copy(o[*c:c.Inc(magicbytes.Len)], Magic)
	copy(o[*c:c.Inc(sha256.Len)], x[:])
}

func (x Type) Decode(b slice.Bytes) (e error) {

	magic := Magic
	if !magicbytes.CheckMagic(b, magic) {
		return magicbytes.WrongMagic(x, b, magic)
	}
	minLen := MinLen
	if len(b) < minLen {
		return magicbytes.TooShort(len(b), minLen, string(magic))
	}
	sc := slice.Cursor(0)
	c := &sc
	_ = c

	return
}
