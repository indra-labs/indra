package wire

import (
	"github.com/Indra-Labs/indra/pkg/sha256"
	"github.com/Indra-Labs/indra/pkg/slice"
	"github.com/Indra-Labs/indra/pkg/types"
)

// Purchase is a message that is sent after first forwarding a Lighting payment
// of an amount corresponding to the number of bytes requested based on the
// price advertised for Exit traffic by a relay.
//
// The Return bytes contain the message header that is prepended to a Session
// message which contains the pair of keys associated with the Session that is
// purchased.
type Purchase struct {
	NBytes uint64
	// Ciphers is a set of 3 symmetric ciphers that are to be used in their
	// given order over the reply message from the service.
	Ciphers [3]sha256.Hash
	types.Onion
}

var (
	PurchaseMagic             = slice.Bytes("prc")
	_             types.Onion = &Purchase{}
)

func (x *Purchase) Inner() types.Onion   { return x.Onion }
func (x *Purchase) Insert(o types.Onion) { x.Onion = o }
func (x *Purchase) Len() int {
	return MagicLen + slice.Uint64Len + x.Onion.Len()
}

func (x *Purchase) Encode(o slice.Bytes, c *slice.Cursor) {
	copy(o[*c:c.Inc(MagicLen)], PurchaseMagic)
	value := slice.NewUint64()
	slice.EncodeUint64(value, x.NBytes)
	x.Onion.Encode(o, c)
}

func (x *Purchase) Decode(b slice.Bytes, c *slice.Cursor) (in interface{},
	e error) {

	magic := PurchaseMagic
	if !CheckMagic(b, magic) {
		return ReturnError(ErrWrongMagic, x, b, magic)
	}

	return
}
