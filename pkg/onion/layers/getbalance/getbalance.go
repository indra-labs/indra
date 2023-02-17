package getbalance

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
	MagicString = "gb"
	Len         = magicbytes.Len + 2*nonce.IDLen +
		3*sha256.Len + nonce.IVLen*3
)

var (
	log               = log2.GetLogger(indra.PathBase)
	check             = log.E.Chk
	Magic             = slice.Bytes(MagicString)
	_     types.Onion = &Layer{}
)

// Layer getbalance messages are a request to return the sats balance of the
// session the message is embedded in.
type Layer struct {
	nonce.ID
	ConfID nonce.ID
	// Ciphers is a set of 3 symmetric ciphers that are to be used in their
	// given order over the reply message from the service.
	Ciphers [3]sha256.Hash
	// Nonces are the nonces to use with the cipher when creating the
	// encryption for the reply message.
	Nonces [3]nonce.IV
	types.Onion
}

// func (x *Layer) String() string {
// 	return spew.Sdump(x.Ciphers, x.Nonces)
// }
func (x *Layer) Insert(o types.Onion) { x.Onion = o }
func (x *Layer) Len() int {
	return Len + x.Onion.Len()
}

func (x *Layer) Encode(b slice.Bytes, c *slice.Cursor) {
	copy(b[*c:c.Inc(magicbytes.Len)], Magic)
	copy(b[*c:c.Inc(nonce.IDLen)], x.ID[:])
	copy(b[*c:c.Inc(nonce.IDLen)], x.ConfID[:])
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
		return magicbytes.TooShort(len(b[*c:]), Len-magicbytes.Len,
			string(Magic))
	}
	copy(x.ID[:], b[*c:c.Inc(nonce.IDLen)])
	copy(x.ConfID[:], b[*c:c.Inc(nonce.IDLen)])
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
