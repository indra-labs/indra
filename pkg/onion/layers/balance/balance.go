package balance

import (
	"git-indra.lan/indra-labs/lnd/lnd/lnwire"
	
	"git-indra.lan/indra-labs/indra"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/magicbytes"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/types"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
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
	copy(b[*c:c.Inc(magicbytes.Len)], Magic)
	copy(b[*c:c.Inc(nonce.IDLen)], x.ID[:])
	copy(b[*c:c.Inc(nonce.IDLen)], x.ConfID[:])
	s := slice.NewUint64()
	slice.EncodeUint64(s, uint64(x.MilliSatoshi))
	copy(b[*c:c.Inc(slice.Uint64Len)], s)
}

func (x *Layer) Decode(b slice.Bytes, c *slice.Cursor) (e error) {
	if len(b[*c:]) < Len-magicbytes.Len {
		return magicbytes.TooShort(len(b[*c:]), Len-magicbytes.Len,
			string(Magic))
	}
	copy(x.ID[:], b[*c:c.Inc(nonce.IDLen)])
	copy(x.ConfID[:], b[*c:c.Inc(nonce.IDLen)])
	x.MilliSatoshi = lnwire.MilliSatoshi(
		slice.DecodeUint64(b[*c:c.Inc(slice.Uint64Len)]))
	return
}
