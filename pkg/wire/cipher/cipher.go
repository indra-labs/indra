package cipher

import (
	"github.com/Indra-Labs/indra/pkg/key/prv"
	"github.com/Indra-Labs/indra/pkg/key/pub"
	"github.com/Indra-Labs/indra/pkg/slice"
	"github.com/Indra-Labs/indra/pkg/types"
	"github.com/Indra-Labs/indra/pkg/wire/magicbytes"
)

// Type cipher delivers a pair of public keys to be used in association with a Return
// specifically in the situation of a node bootstrapping sessions, which doesn't
// have sessions yet.
type Type struct {
	Header, Payload *prv.Key
	types.Onion
}

var (
	Magic              = slice.Bytes("cif")
	MinLen             = magicbytes.Len + pub.KeyLen*2
	_      types.Onion = &Type{}
)

func (x *Type) Inner() types.Onion   { return x.Onion }
func (x *Type) Insert(o types.Onion) { x.Onion = o }
func (x *Type) Len() int {
	return magicbytes.Len + pub.KeyLen + x.Onion.Len()
}

func (x *Type) Encode(o slice.Bytes, c *slice.Cursor) {
	copy(o[*c:c.Inc(magicbytes.Len)], Magic)
	hdr := pub.Derive(x.Header).ToBytes()
	pld := pub.Derive(x.Payload).ToBytes()
	copy(o[c.Inc(1):c.Inc(pub.KeyLen)], hdr[:])
	copy(o[c.Inc(1):c.Inc(pub.KeyLen)], pld[:])
	x.Onion.Encode(o, c)
}

func (x *Type) Decode(b slice.Bytes) (e error) {

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
