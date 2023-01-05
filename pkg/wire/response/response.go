package response

import (
	"github.com/Indra-Labs/indra"
	log2 "github.com/Indra-Labs/indra/pkg/log"
	"github.com/Indra-Labs/indra/pkg/sha256"
	"github.com/Indra-Labs/indra/pkg/slice"
	"github.com/Indra-Labs/indra/pkg/types"
	"github.com/Indra-Labs/indra/pkg/wire/magicbytes"
)

const (
	MagicString = "rs"
	Len         = magicbytes.Len + slice.Uint32Len + sha256.Len
)

var (
	log               = log2.GetLogger(indra.PathBase)
	check             = log.E.Chk
	Magic             = slice.Bytes(MagicString)
	_     types.Onion = &OnionSkin{}
)

// OnionSkin messages are what are carried back via Reverse messages from an Exit.
type OnionSkin struct {
	sha256.Hash
	slice.Bytes
}

func New() *OnionSkin {
	o := OnionSkin{}
	return &o
}

func (x *OnionSkin) Inner() types.Onion   { return nil }
func (x *OnionSkin) Insert(_ types.Onion) {}
func (x *OnionSkin) Len() int             { return Len + len(x.Bytes) }

func (x *OnionSkin) Encode(b slice.Bytes, c *slice.Cursor) {
	copy(b[*c:c.Inc(magicbytes.Len)], Magic)
	copy(b[*c:c.Inc(sha256.Len)], x.Hash[:])
	bytesLen := slice.NewUint32()
	slice.EncodeUint32(bytesLen, len(x.Bytes))
	copy(b[*c:c.Inc(slice.Uint32Len)], bytesLen)
	copy(b[*c:c.Inc(len(x.Bytes))], x.Bytes)
}

func (x *OnionSkin) Decode(b slice.Bytes, c *slice.Cursor) (e error) {
	if len(b[*c:]) < Len-magicbytes.Len {
		return magicbytes.TooShort(len(b[*c:]),
			Len-magicbytes.Len, string(Magic))
	}
	copy(x.Hash[:], b[*c:c.Inc(sha256.Len)])
	responseLen := slice.DecodeUint32(b[*c:c.Inc(slice.Uint32Len)])
	bb := b[*c:c.Inc(responseLen)]
	x.Bytes = bb
	return
}
