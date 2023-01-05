package cipher

import (
	"github.com/Indra-Labs/indra"
	"github.com/Indra-Labs/indra/pkg/key/pub"
	log2 "github.com/Indra-Labs/indra/pkg/log"
	"github.com/Indra-Labs/indra/pkg/slice"
	"github.com/Indra-Labs/indra/pkg/types"
	"github.com/Indra-Labs/indra/pkg/wire/magicbytes"
)

const (
	MagicString = "cf"
	Len         = magicbytes.Len + pub.KeyLen*2
)

var (
	log               = log2.GetLogger(indra.PathBase)
	check             = log.E.Chk
	Magic             = slice.Bytes(MagicString)
	_     types.Onion = &OnionSkin{}
)

// OnionSkin cipher delivers a pair of private keys to be used in association with a
// reply.Type specifically in the situation of a node bootstrapping sessions.
//
// After ~10 seconds these can be purged from the cache as they are otherwise a
// DoS vector buffer flooding.
type OnionSkin struct {
	Header, Payload *pub.Key
	types.Onion
}

func (x *OnionSkin) Inner() types.Onion   { return x.Onion }
func (x *OnionSkin) Insert(o types.Onion) { x.Onion = o }
func (x *OnionSkin) Len() int             { return Len + x.Onion.Len() }

func (x *OnionSkin) Encode(b slice.Bytes, c *slice.Cursor) {
	copy(b[*c:c.Inc(magicbytes.Len)], Magic)
	hdr := x.Header.ToBytes()
	pld := x.Payload.ToBytes()
	copy(b[*c:c.Inc(pub.KeyLen)], hdr[:])
	copy(b[*c:c.Inc(pub.KeyLen)], pld[:])
	x.Onion.Encode(b, c)
}

// Decode unwraps a cipher.OnionSkin message.
func (x *OnionSkin) Decode(b slice.Bytes, c *slice.Cursor) (e error) {
	if len(b[*c:]) < Len-magicbytes.Len {
		return magicbytes.TooShort(len(b[*c:]),
			Len-magicbytes.Len, string(Magic))
	}
	x.Header, e = pub.FromBytes(b[*c:c.Inc(pub.KeyLen)])
	x.Payload, e = pub.FromBytes(b[*c:c.Inc(pub.KeyLen)])
	return
}
