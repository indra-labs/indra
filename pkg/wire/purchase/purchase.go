package purchase

import (
	"github.com/indra-labs/indra"
	log2 "github.com/indra-labs/indra/pkg/log"
	"github.com/indra-labs/indra/pkg/nonce"
	"github.com/indra-labs/indra/pkg/sha256"
	"github.com/indra-labs/indra/pkg/slice"
	"github.com/indra-labs/indra/pkg/types"
	"github.com/indra-labs/indra/pkg/wire/magicbytes"
)

const (
	MagicString = "pc"
	Len         = magicbytes.Len + slice.Uint64Len + sha256.Len*3 +
		nonce.IVLen*3 + nonce.IDLen
)

var (
	log               = log2.GetLogger(indra.PathBase)
	check             = log.E.Chk
	Magic             = slice.Bytes(MagicString)
	_     types.Onion = &OnionSkin{}
)

// OnionSkin purchase is a message that requests a session key, which will
// activate when a payment for it has been done, or it will time out after some
// period to allow unused codes to be flushed.
type OnionSkin struct {
	nonce.ID
	NBytes uint64
	// Ciphers is a set of 3 symmetric ciphers that are to be used in their
	// given order over the reply message from the service.
	Ciphers [3]sha256.Hash
	// Nonces are the nonces to use with the cipher when creating the
	// encryption for the reply message.
	Nonces [3]nonce.IV
	types.Onion
}

func (x *OnionSkin) Inner() types.Onion   { return x.Onion }
func (x *OnionSkin) Insert(o types.Onion) { x.Onion = o }
func (x *OnionSkin) Len() int {
	return Len + x.Onion.Len()
}

func (x *OnionSkin) Encode(b slice.Bytes, c *slice.Cursor) {
	copy(b[*c:c.Inc(magicbytes.Len)], Magic)
	copy(b[*c:c.Inc(nonce.IDLen)], x.ID[:])
	value := slice.NewUint64()
	slice.EncodeUint64(value, x.NBytes)
	copy(b[*c:c.Inc(slice.Uint64Len)], value)
	copy(b[*c:c.Inc(sha256.Len)], x.Ciphers[0][:])
	copy(b[*c:c.Inc(sha256.Len)], x.Ciphers[1][:])
	copy(b[*c:c.Inc(sha256.Len)], x.Ciphers[2][:])
	copy(b[*c:c.Inc(nonce.IVLen)], x.Nonces[0][:])
	copy(b[*c:c.Inc(nonce.IVLen)], x.Nonces[1][:])
	copy(b[*c:c.Inc(nonce.IVLen)], x.Nonces[2][:])
	x.Onion.Encode(b, c)
}

func (x *OnionSkin) Decode(b slice.Bytes, c *slice.Cursor) (e error) {
	if len(b[*c:]) < Len-magicbytes.Len {
		return magicbytes.TooShort(len(b[*c:]),
			Len-magicbytes.Len, MagicString)
	}
	copy(x.ID[:], b[*c:c.Inc(nonce.IDLen)])
	x.NBytes = slice.DecodeUint64(
		b[*c:c.Inc(slice.Uint64Len)])
	for i := range x.Ciphers {
		bytes := b[*c:c.Inc(sha256.Len)]
		copy(x.Ciphers[i][:], bytes)
		bytes.Zero()
	}
	for i := range x.Nonces {
		bytes := b[*c:c.Inc(nonce.IVLen)]
		copy(x.Nonces[i][:], bytes)
		bytes.Zero()
	}
	return
}
