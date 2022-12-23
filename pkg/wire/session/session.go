package session

import (
	"github.com/Indra-Labs/indra"
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

// Type session is a message containing two public keys which identify to a
// relay how to decrypt the header in a Return message, using the HeaderKey, and
// the payload, which uses the PayloadKey. There is two keys in order to prevent
// the Exit node from being able to decrypt the header, but enable it to encrypt
// the payload, and Return relay hops have these key pairs and identify the
// HeaderKey and then know they can unwrap their layer of the payload using the
// PayloadKey.
//
// Clients use the HeaderKey, cloaked, in their messages for the seller relay,
// in the header, and use the PayloadKey as the public key half with ECDH and
// their generated private key which produces the public key that is placed in
// the header.
type Type struct {
	HeaderKey, PayloadKey *pub.Key
	types.Onion
}

var (
	Magic              = slice.Bytes("ses")
	MinLen             = magicbytes.Len + pub.KeyLen*2
	_      types.Onion = &Type{}
)

func (x *Type) Inner() types.Onion   { return x.Onion }
func (x *Type) Insert(o types.Onion) { x.Onion = o }
func (x *Type) Len() int {
	return magicbytes.Len + pub.KeyLen*2 + x.Onion.Len()
}

func (x *Type) Encode(b slice.Bytes, c *slice.Cursor) {
	hdr, pld := x.HeaderKey.ToBytes(), x.PayloadKey.ToBytes()
	copy(b[*c:c.Inc(magicbytes.Len)], Magic)
	copy(b[*c:c.Inc(pub.KeyLen)], hdr[:])
	copy(b[*c:c.Inc(pub.KeyLen)], pld[:])
	x.Onion.Encode(b, c)
}

func (x *Type) Decode(b slice.Bytes, c *slice.Cursor) (e error) {
	if !magicbytes.CheckMagic(b, Magic) {
		return magicbytes.WrongMagic(x, b, Magic)
	}
	if len(b) < MinLen {
		return magicbytes.TooShort(len(b), MinLen, string(Magic))
	}
	if x.HeaderKey, e = pub.FromBytes(b[c.Inc(magicbytes.Len):c.Inc(
		pub.KeyLen)]); check(e) {

		return
	}
	if x.PayloadKey, e = pub.FromBytes(b[*c:c.Inc(pub.KeyLen)]); check(e) {
		return
	}
	return
}
