package wire

import (
	"github.com/Indra-Labs/indra/pkg/key/prv"
	"github.com/Indra-Labs/indra/pkg/key/pub"
	"github.com/Indra-Labs/indra/pkg/slice"
	"github.com/Indra-Labs/indra/pkg/types"
)

// Cipher delivers a public key to be used in association with a Return
// specifically in the situation of a node bootstrapping that doesn't have
// sessions yet. The Forward key will appear in the pre-formed header, but the
// cipher provided to the seller will correspond with
type Cipher struct {
	Header, Payload *prv.Key
	types.Onion
}

var (
	CipherMagic             = slice.Bytes("cif")
	_           types.Onion = &Cipher{}
)

func (x *Cipher) Inner() types.Onion   { return x.Onion }
func (x *Cipher) Insert(o types.Onion) { x.Onion = o }
func (x *Cipher) Len() int {
	return MagicLen + pub.KeyLen + x.Onion.Len()
}

func (x *Cipher) Encode(o slice.Bytes, c *slice.Cursor) {
	copy(o[*c:c.Inc(MagicLen)], CipherMagic)
	hdr, pld := x.Header.ToBytes(), x.Payload.ToBytes()
	copy(o[c.Inc(1):c.Inc(prv.KeyLen)], hdr[:])
	copy(o[c.Inc(1):c.Inc(prv.KeyLen)], pld[:])
	x.Onion.Encode(o, c)
}

func (x *Cipher) Decode(b slice.Bytes, c *slice.Cursor) (in interface{},
	e error) {

	magic := CipherMagic
	if !CheckMagic(b, magic) {
		return ReturnError(ErrWrongMagic, x, b, magic)
	}

	return
}
