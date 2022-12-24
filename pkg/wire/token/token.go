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
	log                     = log2.GetLogger(indra.PathBase)
	check                   = log.E.Chk
	MagicString             = "tk"
	Magic                   = slice.Bytes(MagicString)
	MinLen                  = magicbytes.Len + sha256.Len
	_           types.Onion = Type{}
)

// A Type is a 32 byte value.
type Type sha256.Hash

func (x Type) Inner() types.Onion   { return nil }
func (x Type) Insert(_ types.Onion) {}
func (x Type) Len() int             { return MinLen }

func (x Type) Encode(b slice.Bytes, c *slice.Cursor) {
	copy(b[*c:c.Inc(magicbytes.Len)], Magic)
	copy(b[*c:c.Inc(sha256.Len)], x[:])
}

func (x Type) Decode(b slice.Bytes, c *slice.Cursor) (e error) {
	if len(b[*c:]) < MinLen {
		return magicbytes.TooShort(len(b[*c:]), MinLen, string(Magic))
	}
	copy(x[:], b[c.Inc(magicbytes.Len):c.Inc(sha256.Len)])
	return
}
