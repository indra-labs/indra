package session

import (
	"github.com/indra-labs/indra"
	"github.com/indra-labs/indra/pkg/key/pub"
	log2 "github.com/indra-labs/indra/pkg/log"
	"github.com/indra-labs/indra/pkg/nonce"
	"github.com/indra-labs/indra/pkg/slice"
	"github.com/indra-labs/indra/pkg/types"
	"github.com/indra-labs/indra/pkg/wire/magicbytes"
)

const (
	MagicString = "ss"
	Len         = magicbytes.Len + nonce.IDLen + pub.KeyLen*2
)

var (
	log               = log2.GetLogger(indra.PathBase)
	check             = log.E.Chk
	Magic             = slice.Bytes(MagicString)
	_     types.Onion = &OnionSkin{}
)

// OnionSkin session is a message containing two public keys which identify to a
// relay how to decrypt the header in a Reverse message, using the HeaderKey, and
// the payload, which uses the PayloadKey. There is two keys in order to prevent
// the Exit node from being able to decrypt the header, but enable it to encrypt
// the payload, and Reverse relay hops have these key pairs and identify the
// HeaderKey and then know they can unwrap their layer of the payload using the
// PayloadKey.
//
// Clients use the HeaderKey, cloaked, in their messages for the seller relay,
// in the header, and use the PayloadKey as the public key half with ECDH and
// their generated private key which produces the public key that is placed in
// the header associated with the payload, using the same nonce.IV as well.
type OnionSkin struct {
	nonce.ID
	HeaderKey, PayloadKey *pub.Key
	types.Onion
}

func (x *OnionSkin) Inner() types.Onion   { return x.Onion }
func (x *OnionSkin) Insert(o types.Onion) { x.Onion = o }
func (x *OnionSkin) Len() int             { return Len + x.Onion.Len() }
func (x *OnionSkin) Encode(b slice.Bytes, c *slice.Cursor) {
	hdr, pld := x.HeaderKey.ToBytes(), x.PayloadKey.ToBytes()
	copy(b[*c:c.Inc(magicbytes.Len)], Magic)
	copy(b[*c:c.Inc(nonce.IDLen)], x.ID[:])
	copy(b[*c:c.Inc(pub.KeyLen)], hdr[:])
	copy(b[*c:c.Inc(pub.KeyLen)], pld[:])
	x.Onion.Encode(b, c)
}
func (x *OnionSkin) Decode(b slice.Bytes, c *slice.Cursor) (e error) {
	if len(b[*c:]) < Len-magicbytes.Len {
		return magicbytes.TooShort(len(b[*c:]), Len-magicbytes.Len, string(Magic))
	}
	copy(x.ID[:], b[*c:c.Inc(nonce.IDLen)])
	if x.HeaderKey, e = pub.FromBytes(b[*c:c.Inc(pub.KeyLen)]); check(e) {
		return
	}
	if x.PayloadKey, e = pub.FromBytes(b[*c:c.Inc(pub.KeyLen)]); check(e) {
		return
	}
	return
}
