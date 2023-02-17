package delay

import (
	"time"
	
	"git-indra.lan/indra-labs/indra"
	"git-indra.lan/indra-labs/indra/pkg/onion/magicbytes"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/types"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

const (
	MagicString = "dl"
	Len         = magicbytes.Len + slice.Uint64Len
)

var (
	log               = log2.GetLogger(indra.PathBase)
	check             = log.E.Chk
	Magic             = slice.Bytes(MagicString)
	_     types.Onion = &Layer{}
)

// A Layer delay is a message to hold for a period of time before relaying.
type Layer struct {
	time.Duration
	types.Onion
}

func (x *Layer) Insert(_ types.Onion) {}
func (x *Layer) Len() int             { return Len }

func (x *Layer) Encode(b slice.Bytes, c *slice.Cursor) {
	copy(b[*c:c.Inc(magicbytes.Len)], Magic)
	slice.EncodeUint64(b[*c:c.Inc(slice.Uint64Len)], uint64(x.Duration))
	x.Onion.Encode(b, c)
}

func (x *Layer) Decode(b slice.Bytes, c *slice.Cursor) (e error) {
	if len(b[*c:]) < Len-magicbytes.Len {
		return magicbytes.TooShort(len(b[*c:]), Len-magicbytes.Len,
			string(Magic))
	}
	x.Duration = time.Duration(
		slice.DecodeUint64(b[*c:c.Inc(slice.Uint64Len)]))
	return
}
