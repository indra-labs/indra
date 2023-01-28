package response

import (
	"github.com/indra-labs/indra"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/crypto/sha256"
	"github.com/indra-labs/indra/pkg/onion/layers/magicbytes"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/types"
	"github.com/indra-labs/indra/pkg/util/slice"
)

const (
	MagicString = "rs"
	Len         = magicbytes.Len + slice.Uint32Len + sha256.Len +
		nonce.IDLen + 1
)

var (
	log               = log2.GetLogger(indra.PathBase)
	check             = log.E.Chk
	Magic             = slice.Bytes(MagicString)
	_     types.Onion = &Layer{}
)

// Layer messages are what are carried back via Reverse messages from an Exit.
type Layer struct {
	sha256.Hash
	nonce.ID
	Load byte
	slice.Bytes
}

func New() *Layer {
	o := Layer{}
	return &o
}

func (x *Layer) Inner() types.Onion   { return nil }
func (x *Layer) Insert(_ types.Onion) {}
func (x *Layer) Len() int             { return Len + len(x.Bytes) }
func (x *Layer) Encode(b slice.Bytes, c *slice.Cursor) {
	copy(b[*c:c.Inc(magicbytes.Len)], Magic)
	copy(b[*c:c.Inc(sha256.Len)], x.Hash[:])
	copy(b[*c:c.Inc(nonce.IDLen)], x.ID[:])
	b[*c] = x.Load
	c.Inc(1)
	bytesLen := slice.NewUint32()
	slice.EncodeUint32(bytesLen, len(x.Bytes))
	copy(b[*c:c.Inc(slice.Uint32Len)], bytesLen)
	copy(b[*c:c.Inc(len(x.Bytes))], x.Bytes)
}
func (x *Layer) Decode(b slice.Bytes, c *slice.Cursor) (e error) {
	if len(b[*c:]) < Len-magicbytes.Len {
		return magicbytes.TooShort(len(b[*c:]),
			Len-magicbytes.Len, string(Magic))
	}
	copy(x.Hash[:], b[*c:c.Inc(sha256.Len)])
	copy(x.ID[:], b[*c:c.Inc(nonce.IDLen)])
	x.Load = b[*c]
	c.Inc(1)
	responseLen := slice.DecodeUint32(b[*c:c.Inc(slice.Uint32Len)])
	bb := b[*c:c.Inc(responseLen)]
	x.Bytes = bb
	return
}
