package response

import (
	"git-indra.lan/indra-labs/indra"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/messages/magicbytes"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/splice"
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
	splice.Splice(b, c).
		Magic(Magic).
		ID(x.ID).
		Uint16(x.Port).
		Byte(x.Load).
		Bytes(x.Bytes)
}
func (x *Layer) Decode(b slice.Bytes, c *slice.Cursor) (e error) {
	if len(b[*c:]) < Len-magicbytes.Len {
		return magicbytes.TooShort(len(b[*c:]),
			Len-magicbytes.Len, string(Magic))
	}
	splice.Splice(b, c).
		ReadID(&x.ID).
		ReadUint16(&x.Port).
		ReadByte(&x.Load).
		ReadBytes(&x.Bytes)
	return
}
