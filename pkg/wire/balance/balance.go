package balance

import (
	"github.com/indra-labs/indra"
	"github.com/indra-labs/indra/pkg/lnd/lnwire"
	log2 "github.com/indra-labs/indra/pkg/log"
	"github.com/indra-labs/indra/pkg/nonce"
	"github.com/indra-labs/indra/pkg/slice"
	"github.com/indra-labs/indra/pkg/types"
	"github.com/indra-labs/indra/pkg/wire/magicbytes"
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
	_     types.Onion = &OnionSkin{}
)

// OnionSkin balance messages are the response replying to a GetBalance message.
type OnionSkin struct {
	nonce.ID
	lnwire.MilliSatoshi
}

func (x *OnionSkin) Inner() types.Onion   { return nil }
func (x *OnionSkin) Insert(o types.Onion) {}
func (x *OnionSkin) Len() int {
	return Len
}

func (x *OnionSkin) Encode(b slice.Bytes, c *slice.Cursor) {
	copy(b[*c:c.Inc(magicbytes.Len)], Magic)
	copy(b[*c:c.Inc(nonce.IDLen)], x.ID[:])
	s := slice.NewUint64()
	slice.EncodeUint64(s, uint64(x.MilliSatoshi))
	copy(b[*c:slice.Uint64Len], s)
}

func (x *OnionSkin) Decode(b slice.Bytes, c *slice.Cursor) (e error) {
	if len(b[*c:]) < Len-magicbytes.Len {
		return magicbytes.TooShort(len(b[*c:]), Len-magicbytes.Len,
			string(Magic))
	}
	copy(x.ID[:], b[*c:nonce.IDLen])
	x.MilliSatoshi = lnwire.MilliSatoshi(
		slice.DecodeUint64(b[*c:c.Inc(slice.Uint64Len)]))
	return
}
