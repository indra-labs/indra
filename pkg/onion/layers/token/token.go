package token

import (
	"github.com/indra-labs/indra"
	"github.com/indra-labs/indra/pkg/crypto/sha256"
	"github.com/indra-labs/indra/pkg/onion/layers/magicbytes"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/types"
	"github.com/indra-labs/indra/pkg/util/slice"
)

var (
	log                     = log2.GetLogger(indra.PathBase)
	check                   = log.E.Chk
	MagicString             = "tk"
	Magic                   = slice.Bytes(MagicString)
	MinLen                  = magicbytes.Len + sha256.Len
	_           types.Onion = &Layer{}
)

// A Layer is a 32 byte value.
type Layer sha256.Hash

func NewOnionSkin() *Layer {
	var os sha256.Hash
	return (*Layer)(&os)
}

func (x *Layer) Inner() types.Onion   { return nil }
func (x *Layer) Insert(_ types.Onion) {}
func (x *Layer) Len() int             { return MinLen }
func (x *Layer) Encode(b slice.Bytes, c *slice.Cursor) {
	copy(b[*c:c.Inc(magicbytes.Len)], Magic)
	copy(b[*c:c.Inc(sha256.Len)], x[:sha256.Len])
}
func (x *Layer) Decode(b slice.Bytes, c *slice.Cursor) (e error) {
	if len(b[*c:]) < MinLen-magicbytes.Len {
		return magicbytes.TooShort(len(b[*c:]), MinLen-magicbytes.Len,
			string(Magic))
	}
	copy(x[:], b[*c:c.Inc(sha256.Len)])
	return
}
