package confirmation

import (
	"github.com/Indra-Labs/indra"
	"github.com/Indra-Labs/indra/pkg/nonce"
	"github.com/Indra-Labs/indra/pkg/slice"
	"github.com/Indra-Labs/indra/pkg/types"
	"github.com/Indra-Labs/indra/pkg/wire/magicbytes"
	log2 "github.com/cybriq/proc/pkg/log"
)

var (
	log                     = log2.GetLogger(indra.PathBase)
	check                   = log.E.Chk
	MagicString             = "cnf"
	Magic                   = slice.Bytes(MagicString)
	MinLen                  = magicbytes.Len + nonce.IDLen
	_           types.Onion = &Type{}
)

// Type confirmation is an encryption layer for messages returned to the client
// on the inside of an onion, for Ping and Ciphers messages, providing a
// confirmation of the transit of the onion through its encoded route.
//
// It is encrypted because otherwise internal identifiers could be leaked and
// potentially reveal something about the entropy of a client/relay.
//
// In order to speed up recognition, the key of the table of pending Ping and
// Ciphers messages will include the last hop that will deliver this layer of the
// onion - there can be more than one up in the air at a time, but they are
// randomly selected, so they will generally be a much smaller subset versus the
// current full set of Session s currently open.
type Type struct {
	nonce.ID
}

func (x *Type) Inner() types.Onion   { return nil }
func (x *Type) Insert(o types.Onion) {}
func (x *Type) Len() int             { return MinLen }

func (x *Type) Encode(b slice.Bytes, c *slice.Cursor) {
	copy(b[*c:c.Inc(magicbytes.Len)], Magic)
	// Copy in the ID.
	copy(b[*c:c.Inc(nonce.IDLen)], x.ID[:])
}

func (x *Type) Decode(b slice.Bytes, c *slice.Cursor) (e error) {
	if len(b[*c:]) < MinLen {
		return magicbytes.TooShort(len(b), MinLen, string(Magic))
	}
	copy(x.ID[:], b[*c:c.Inc(nonce.IDLen)])
	return
}
