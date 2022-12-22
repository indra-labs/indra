package wire

import (
	"github.com/Indra-Labs/indra/pkg/nonce"
	"github.com/Indra-Labs/indra/pkg/slice"
	"github.com/Indra-Labs/indra/pkg/types"
)

// Confirmation is an encryption layer for messages returned to the client on
// the inside of an onion, for Ping and Cipher messages, providing a
// confirmation of the transit of the onion through its encoded route.
//
// It is encrypted because otherwise internal identifiers could be leaked and
// potentially reveal something about the entropy of a client/relay.
//
// In order to speed up recognition, the key of the table of pending Ping and
// Cipher messages will include the last hop that will deliver this layer of the
// onion - there can be more than one up in the air at a time, but they are
// randomly selected, so they will generally be a much smaller subset versus the
// current full set of Session s currently open.
type Confirmation struct {
	nonce.ID
}

var (
	ConfirmationMagic             = slice.Bytes("cnf")
	_                 types.Onion = &Confirmation{}
)

func (x *Confirmation) Inner() types.Onion   { return nil }
func (x *Confirmation) Insert(o types.Onion) {}
func (x *Confirmation) Len() int {
	return MagicLen + nonce.IDLen
}

func (x *Confirmation) Encode(o slice.Bytes, c *slice.Cursor) {
	copy(o[*c:c.Inc(MagicLen)], ConfirmationMagic)
	// Copy in the ID.
	copy(o[*c:c.Inc(nonce.IDLen)], x.ID[:])
}

func (x *Confirmation) Decode(b slice.Bytes, c *slice.Cursor) (in interface{},
	e error) {

	magic := ConfirmationMagic
	if !CheckMagic(b, magic) {
		return ReturnError(ErrWrongMagic, x, b, magic)
	}

	return
}
