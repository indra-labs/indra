package balance

import (
	"git-indra.lan/indra-labs/lnd/lnd/lnwire"
	
	"git-indra.lan/indra-labs/indra"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/relay/messages/magicbytes"
	"git-indra.lan/indra-labs/indra/pkg/relay/types"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
	"git-indra.lan/indra-labs/indra/pkg/util/splice"
)

const (
	MagicString = "ba"
	Len         = magicbytes.Len + 2*nonce.IDLen +
		slice.Uint64Len
)

var (
	log               = log2.GetLogger(indra.PathBase)
	check             = log.E.Chk
	Magic             = slice.Bytes(MagicString)
	_     types.Onion = &Layer{}
)

// Layer balance messages are the response replying to a GetBalance message. The
// ID is a random value that quickly identifies to the client which request it
// relates to for the callback.
type Layer struct {
	nonce.ID
	ConfID nonce.ID
	lnwire.MilliSatoshi
}

func (x *Layer) Insert(o types.Onion) {}
func (x *Layer) Len() int             { return Len }

func (x *Layer) Encode(b slice.Bytes, c *slice.Cursor) {
	splice.Splice(b, c).
		Magic(Magic).
		ID(x.ID).
		ID(x.ConfID).
		Uint64(uint64(x.MilliSatoshi))
}

func (x *Layer) Decode(b slice.Bytes, c *slice.Cursor) (e error) {
	if len(b[*c:]) < Len-magicbytes.Len {
		return magicbytes.TooShort(len(b[*c:]), Len-magicbytes.Len,
			string(Magic))
	}
	splice.Splice(b, c).
		ReadID(&x.ID).
		ReadID(&x.ConfID).
		ReadMilliSatoshi(&x.MilliSatoshi)
	return
}
