package balance

import (
	"github.com/indra-labs/indra"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/onion/layers/magicbytes"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/types"
	"github.com/indra-labs/indra/pkg/util/slice"
	"github.com/indra-labs/lnd/lnd/lnwire"
)

const (
	MagicString = "ba"
	Len         = magicbytes.Len + nonce.IDLen +
		slice.Uint64Len
)

var (
	log               = log2.GetLogger(indra.PathBase)
	check             = log.E.Chk
	Magic             = slice.Bytes(MagicString)
	_     types.Onion = &Layer{}
)

// Layer balance messages are the response replying to a GetBalance message.
type Layer struct {
	nonce.ID
	lnwire.MilliSatoshi
}

func (x *Layer) Inner() types.Onion   { return nil }
func (x *Layer) Insert(o types.Onion) {}
func (x *Layer) Len() int {
	return Len
}

func (x *Layer) Encode(b slice.Bytes, c *slice.Cursor) {
	copy(b[*c:c.Inc(magicbytes.Len)], Magic)
	copy(b[*c:c.Inc(nonce.IDLen)], x.ID[:])
	s := slice.NewUint64()
	slice.EncodeUint64(s, uint64(x.MilliSatoshi))
	copy(b[*c:slice.Uint64Len], s)
}

func (x *Layer) Decode(b slice.Bytes, c *slice.Cursor) (e error) {
	if len(b[*c:]) < Len-magicbytes.Len {
		return magicbytes.TooShort(len(b[*c:]), Len-magicbytes.Len,
			string(Magic))
	}
	copy(x.ID[:], b[*c:nonce.IDLen])
	x.MilliSatoshi = lnwire.MilliSatoshi(
		slice.DecodeUint64(b[*c:c.Inc(slice.Uint64Len)]))
	return
}
