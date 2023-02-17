package response

import (
	"git-indra.lan/indra-labs/indra"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/onion/magicbytes"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/types"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

const (
	MagicString = "rs"
	Len         = magicbytes.Len + slice.Uint32Len + slice.Uint16Len +
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
	nonce.ID
	Port uint16
	Load byte
	slice.Bytes
}

func New() *Layer {
	o := Layer{}
	return &o
}

func (x *Layer) Insert(_ types.Onion) {}
func (x *Layer) Len() int             { return Len + len(x.Bytes) }
func (x *Layer) Encode(b slice.Bytes, c *slice.Cursor) {
	copy(b[*c:c.Inc(magicbytes.Len)], Magic)
	copy(b[*c:c.Inc(nonce.IDLen)], x.ID[:])
	port := slice.NewUint16()
	slice.EncodeUint16(port, int(x.Port))
	copy(b[*c:c.Inc(slice.Uint16Len)], port)
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
	copy(x.ID[:], b[*c:c.Inc(nonce.IDLen)])
	port := slice.DecodeUint16(b[*c:c.Inc(slice.Uint16Len)])
	x.Port = uint16(port)
	x.Load = b[*c]
	c.Inc(1)
	responseLen := slice.DecodeUint32(b[*c:c.Inc(slice.Uint32Len)])
	bb := b[*c:c.Inc(responseLen)]
	x.Bytes = bb
	return
}
