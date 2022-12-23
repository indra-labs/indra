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

// Type cipher delivers a pair of private keys to be used in association with a
// reply.Type specifically in the situation of a node bootstrapping sessions.
//
// After ~10 seconds these can be purged from the cache as they are otherwise a
// DoS vector buffer flooding.
//
// The Decode function wipes the original message data for security as the
// private keys inside it are no longer needed and any secret should only have
// one storage, so it doesn't appear in any GC later.
type Type struct {
	Header, Payload *pub.Key
	types.Onion
}

var (
	Magic              = slice.Bytes("cif")
	MinLen             = magicbytes.Len + prv.KeyLen*2
	_      types.Onion = &Type{}
)

func (x *Type) Inner() types.Onion   { return x.Onion }
func (x *Type) Insert(o types.Onion) { x.Onion = o }
func (x *Type) Len() int {
	return magicbytes.Len + pub.KeyLen + x.Onion.Len()
}

func (x *Type) Encode(b slice.Bytes, c *slice.Cursor) {
	copy(b[*c:c.Inc(magicbytes.Len)], Magic)
	hdr := x.Header.ToBytes()
	pld := x.Payload.ToBytes()
	copy(b[c.Inc(1):c.Inc(pub.KeyLen)], hdr[:])
	copy(b[c.Inc(1):c.Inc(pub.KeyLen)], pld[:])
	x.Onion.Encode(b, c)
}

// Decode unwraps a cipher.Type message.
func (x *Type) Decode(b slice.Bytes, c *slice.Cursor) (e error) {
	if !magicbytes.CheckMagic(b, Magic) {
		return magicbytes.WrongMagic(x, b, Magic)
	}
	if len(b) < MinLen {
		return magicbytes.TooShort(len(b), MinLen, string(Magic))
	}
	start := c.Inc(magicbytes.Len)
	x.Header, e = pub.FromBytes(b[start:c.Inc(pub.KeyLen)])
	x.Payload, e = pub.FromBytes(b[*c:c.Inc(pub.KeyLen)])
	return
}
