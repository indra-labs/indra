package confirm

import (
	"fmt"

	"github.com/indra-labs/indra"
	log2 "github.com/indra-labs/indra/pkg/log"
	"github.com/indra-labs/indra/pkg/nonce"
	"github.com/indra-labs/indra/pkg/slice"
	"github.com/indra-labs/indra/pkg/types"
	"github.com/indra-labs/indra/pkg/wire/magicbytes"
)

const (
	MagicString = "cn"
	Len         = magicbytes.Len + nonce.IDLen
)

var (
	log               = log2.GetLogger(indra.PathBase)
	check             = log.E.Chk
	Magic             = slice.Bytes(MagicString)
	_     types.Onion = &OnionSkin{}
)

// OnionSkin confirmation is an encryption layer for messages returned to the
// client on the inside of an onion, for Ping and Ciphers messages, providing a
// confirmation of the transit of the onion through its encoded route.
//
// It is encrypted because otherwise internal identifiers could be leaked and
// potentially reveal something about the entropy of a client/relay.
//
// In order to speed up recognition, the key of the table of pending Ping and
// Ciphers messages will include the last hop that will deliver this layer of
// the onion - there can be more than one up in the air at a time, but they are
// randomly selected, so they will generally be a much smaller subset versus the
// current full set of Session s currently open.
type OnionSkin struct {
	nonce.ID
}

func (x *OnionSkin) String() string {
	return fmt.Sprintf("\n\tnonce: %x\n",
		x.ID)
}

func (x *OnionSkin) Inner() types.Onion   { return nil }
func (x *OnionSkin) Insert(o types.Onion) {}
func (x *OnionSkin) Len() int             { return Len }
func (x *OnionSkin) Encode(b slice.Bytes, c *slice.Cursor) {
	copy(b[*c:c.Inc(magicbytes.Len)], Magic)
	// Copy in the ID.
	copy(b[*c:c.Inc(nonce.IDLen)], x.ID[:])
}
func (x *OnionSkin) Decode(b slice.Bytes, c *slice.Cursor) (e error) {
	if len(b[*c:]) < Len-magicbytes.Len {
		return magicbytes.TooShort(len(b[*c:]),
			Len-magicbytes.Len, string(Magic))
	}
	copy(x.ID[:], b[*c:c.Inc(nonce.IDLen)])
	return
}
