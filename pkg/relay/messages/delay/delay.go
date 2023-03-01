package delay

import (
	"time"
	
	"git-indra.lan/indra-labs/indra"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/relay/messages/magicbytes"
	"git-indra.lan/indra-labs/indra/pkg/relay/types"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
	"git-indra.lan/indra-labs/indra/pkg/util/splice"
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
	splice.Splice(b, c).
		Magic(Magic).
		Uint64(uint64(x.Duration))
	x.Onion.Encode(b, c)
}

func (x *Layer) Decode(b slice.Bytes, c *slice.Cursor) (e error) {
	if len(b[*c:]) < Len-magicbytes.Len {
		return magicbytes.TooShort(len(b[*c:]),
			Len-magicbytes.Len, string(Magic))
	}
	splice.Splice(b, c).ReadDuration(&x.Duration)
	return
}
