package dxresponse

import (
	"git-indra.lan/indra-labs/indra"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/onion/magicbytes"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/splice"
	"git-indra.lan/indra-labs/indra/pkg/types"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

const (
	MagicString = "dr"
	Len         = magicbytes.Len + nonce.IDLen + 1
)

var (
	log               = log2.GetLogger(indra.PathBase)
	check             = log.E.Chk
	Magic             = slice.Bytes(MagicString)
	_     types.Onion = &Layer{}
)

// Layer confirmation is an encryption crypt for messages returned to the
// client on the inside of an onion, for Ping and Ciphers messages, providing a
// confirmation of the transit of the onion through its encoded route.
//
// It is encrypted because otherwise internal identifiers could be leaked and
// potentially reveal something about the entropy of a client/relay.
//
// In order to speed up recognition, the key of the table of pending Ping and
// Ciphers messages will include the last hop that will deliver this crypt of
// the onion - there can be more than one up in the air at a time, but they are
// randomly selected, so they will generally be a much smaller subset versus the
// current full set of Session s currently open.
type Layer struct {
	nonce.ID
	Load byte
}

// func (x *Layer) String() string {
// 	return fmt.Sprintf("\n\tnonce: %x\n",
// 		x.ID)
// }

func (x *Layer) Insert(o types.Onion) {}
func (x *Layer) Len() int             { return Len }
func (x *Layer) Encode(b slice.Bytes, c *slice.Cursor) {
	splice.Splice(b, c).
		Magic(Magic).
		ID(x.ID).
		Byte(x.Load)
}
func (x *Layer) Decode(b slice.Bytes, c *slice.Cursor) (e error) {
	if len(b[*c:]) < Len-magicbytes.Len {
		return magicbytes.TooShort(len(b[*c:]),
			Len-magicbytes.Len, string(Magic))
	}
	splice.Splice(b, c).
		ReadID(&x.ID).
		ReadByte(&x.Load)
	return
}
