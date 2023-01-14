package token

import (
	"github.com/indra-labs/indra"
	log2 "github.com/indra-labs/indra/pkg/log"
	"github.com/indra-labs/indra/pkg/sha256"
	"github.com/indra-labs/indra/pkg/slice"
	"github.com/indra-labs/indra/pkg/types"
	"github.com/indra-labs/indra/pkg/wire/magicbytes"
)

var (
	log                     = log2.GetLogger(indra.PathBase)
	check                   = log.E.Chk
	MagicString             = "tk"
	Magic                   = slice.Bytes(MagicString)
	MinLen                  = magicbytes.Len + sha256.Len
	_           types.Onion = &OnionSkin{}
)

// A OnionSkin is a 32 byte value.
type OnionSkin sha256.Hash

func NewOnionSkin() *OnionSkin {
	var os sha256.Hash
	return (*OnionSkin)(&os)
}

func (x *OnionSkin) Inner() types.Onion   { return nil }
func (x *OnionSkin) Insert(_ types.Onion) {}
func (x *OnionSkin) Len() int             { return MinLen }
func (x *OnionSkin) Encode(b slice.Bytes, c *slice.Cursor) {
	copy(b[*c:c.Inc(magicbytes.Len)], Magic)
	copy(b[*c:c.Inc(sha256.Len)], x[:sha256.Len])
}
func (x *OnionSkin) Decode(b slice.Bytes, c *slice.Cursor) (e error) {
	if len(b[*c:]) < MinLen-magicbytes.Len {
		return magicbytes.TooShort(len(b[*c:]), MinLen-magicbytes.Len,
			string(Magic))
	}
	copy(x[:], b[*c:c.Inc(sha256.Len)])
	return
}
