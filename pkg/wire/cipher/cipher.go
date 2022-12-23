package cipher

import (
	"github.com/Indra-Labs/indra"
	"github.com/Indra-Labs/indra/pkg/key/prv"
	"github.com/Indra-Labs/indra/pkg/key/pub"
	"github.com/Indra-Labs/indra/pkg/slice"
	"github.com/Indra-Labs/indra/pkg/types"
	"github.com/Indra-Labs/indra/pkg/wire/magicbytes"
	log2 "github.com/cybriq/proc/pkg/log"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

// Type cipher delivers a pair of public keys to be used in association with a
// Return specifically in the situation of a node bootstrapping sessions, which
// doesn't have sessions yet.
type Type struct {
	Header, Payload         pub.Bytes
	PrivHeader, PrivPayload *prv.Key
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
	hdr := pub.Derive(x.PrivHeader).ToBytes()
	pld := pub.Derive(x.PrivPayload).ToBytes()
	copy(o[c.Inc(1):c.Inc(pub.KeyLen)], hdr[:])
	copy(o[c.Inc(1):c.Inc(pub.KeyLen)], pld[:])
	x.Onion.Encode(o, c)
}

func (x *Type) Decode(b slice.Bytes, c *slice.Cursor) (e error) {

	if !magicbytes.CheckMagic(b, Magic) {
		return magicbytes.WrongMagic(x, b, Magic)
	}
	if len(b) < MinLen {
		return magicbytes.TooShort(len(b), MinLen, string(Magic))
	}
	var header, payload *pub.Key
	if header, e = pub.FromBytes(
		b[c.Inc(magicbytes.Len):c.Inc(pub.KeyLen)]); check(e) {

		// this ensures the key is on the curve.
		return
	}
	if payload, e = pub.FromBytes(b[*c:c.Inc(pub.KeyLen)]); check(e) {
		// this ensures the key is on the curve.
		return
	}
	x.Header, x.Payload = header.ToBytes(), payload.ToBytes()
	return
}
