package wire

import (
	"github.com/Indra-Labs/indra/pkg/key/pub"
	"github.com/Indra-Labs/indra/pkg/slice"
	"github.com/Indra-Labs/indra/pkg/types"
)

// Session is a message containing two public keys which identify to a relay how
// to decrypt the header in a Return message, using the HeaderKey, and the
// payload, which uses the PayloadKey. There is two keys in order to prevent the
// Exit node from being able to decrypt the header, but enable it to encrypt the
// payload, and Return relay hops have these key pairs and identify the
// HeaderKey and then know they can unwrap their layer of the payload using the
// PayloadKey.
//
// Clients use the HeaderKey, cloaked, in their messages for the seller relay,
// in the header, and use the PayloadKey as the public key half with ECDH and
// their generated private key which produces the public key that is placed in
// the header.
type Session struct {
	HeaderKey  *pub.Key
	PayloadKey *pub.Key
	types.Onion
}

var (
	SessionMagic             = slice.Bytes("ses")
	_            types.Onion = &Session{}
)

func (x *Session) Inner() types.Onion   { return x.Onion }
func (x *Session) Insert(o types.Onion) { x.Onion = o }
func (x *Session) Len() int {
	return MagicLen + pub.KeyLen*2 + x.Onion.Len()
}

func (x *Session) Encode(o slice.Bytes, c *slice.Cursor) {
	hdr, pld := x.HeaderKey.ToBytes(), x.PayloadKey.ToBytes()
	copy(o[*c:c.Inc(MagicLen)], SessionMagic)
	copy(o[*c:c.Inc(pub.KeyLen)], hdr[:])
	copy(o[*c:c.Inc(pub.KeyLen)], pld[:])
	x.Onion.Encode(o, c)
}

func (x *Session) Decode(b slice.Bytes, c *slice.Cursor) (in interface{},
	e error) {

	magic := SessionMagic
	if !CheckMagic(b, magic) {
		return ReturnError(ErrWrongMagic, x, b, magic)
	}

	return
}
