package wire

import (
	"github.com/Indra-Labs/indra/pkg/sha256"
	"github.com/Indra-Labs/indra/pkg/slice"
	"github.com/Indra-Labs/indra/pkg/types"
)

// A Token is a 32 byte value.
type Token sha256.Hash

var (
	TokenMagic             = slice.Bytes("tok")
	_          types.Onion = Token{}
)

func (x Token) Inner() types.Onion   { return nil }
func (x Token) Insert(_ types.Onion) {}
func (x Token) Len() int             { return MagicLen + sha256.Len }

func (x Token) Encode(o slice.Bytes, c *slice.Cursor) {
	copy(o[*c:c.Inc(MagicLen)], TokenMagic)
	copy(o[*c:c.Inc(sha256.Len)], x[:])
}

func (x Token) Decode(b slice.Bytes, c *slice.Cursor) (in interface{},
	e error) {

	magic := TokenMagic
	if !CheckMagic(b, magic) {
		return ReturnError(ErrWrongMagic, x, b, magic)
	}

	return
}
