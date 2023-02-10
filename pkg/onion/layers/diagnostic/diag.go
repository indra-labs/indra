package diag

import (
	"git-indra.lan/indra-labs/indra"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/magicbytes"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/types"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

const (
	MagicString = "dx"
	Len         = magicbytes.Len + slice.Uint16Len + 3*sha256.Len +
		slice.Uint32Len + nonce.IVLen*3 + nonce.IDLen
)

var (
	log               = log2.GetLogger(indra.PathBase)
	check             = log.E.Chk
	Magic             = slice.Bytes(MagicString)
	_     types.Onion = &Layer{}
)

// Layer diag messages are just a set of ciphers and nonces that are followed by
// 3 reverse/crypt layers that are followed by confirmation layer with the ID
// from which they were sent and the relay's load level byte. They are used to
// handle message delivery timeouts to diagnose the state of all relays in a
// given circuit in order to determine which of the hops failed after such a
// timeout.
type Layer struct {
	// Ciphers is a set of 3 symmetric ciphers that are to be used in their
	// given order over the reply message from the service.
	Ciphers [3]sha256.Hash
	// Nonces are the nonces to use with the cipher when creating the encryption
	// for the reply message.
	Nonces [3]nonce.IV
	types.Onion
}

//
// func (x *Layer) String() string {
// 	return spew.Sdump(x.Ciphers, x.Nonces)
// }

func (x *Layer) Inner() types.Onion   { return x.Onion }
func (x *Layer) Insert(o types.Onion) { x.Onion = o }
func (x *Layer) Len() int {
	return Len + x.Onion.Len()
}

func (x *Layer) Encode(b slice.Bytes, c *slice.Cursor) {
	copy(b[*c:c.Inc(magicbytes.Len)], Magic)
	copy(b[*c:c.Inc(sha256.Len)], x.Ciphers[0][:])
	copy(b[*c:c.Inc(sha256.Len)], x.Ciphers[1][:])
	copy(b[*c:c.Inc(sha256.Len)], x.Ciphers[2][:])
	copy(b[*c:c.Inc(nonce.IVLen)], x.Nonces[0][:])
	copy(b[*c:c.Inc(nonce.IVLen)], x.Nonces[1][:])
	copy(b[*c:c.Inc(nonce.IVLen)], x.Nonces[2][:])
	x.Onion.Encode(b, c)
}

func (x *Layer) Decode(b slice.Bytes, c *slice.Cursor) (e error) {
	if len(b[*c:]) < Len-magicbytes.Len {
		return magicbytes.TooShort(len(b[*c:]), Len-magicbytes.Len, string(Magic))
	}
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
