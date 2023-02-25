package getbalance

import (
	"git-indra.lan/indra-labs/indra"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/messages/magicbytes"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/splice"
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
	splice.Splice(b, c).
		Magic(Magic).
		ID(x.ID).ID(x.ConfID).
		Hash(x.Ciphers[0]).Hash(x.Ciphers[1]).Hash(x.Ciphers[2]).
		IV(x.Nonces[0]).IV(x.Nonces[1]).IV(x.Nonces[2])
	x.Onion.Encode(b, c)
}

func (x *Layer) Decode(b slice.Bytes, c *slice.Cursor) (e error) {
	if len(b[*c:]) < Len-magicbytes.Len {
		return magicbytes.TooShort(len(b[*c:]), Len-magicbytes.Len,
			string(Magic))
	}
	
	splice.Splice(b, c).
		ReadID(&x.ID).ReadID(&x.ConfID).
		ReadHash(&x.Ciphers[0]).ReadHash(&x.Ciphers[1]).ReadHash(&x.Ciphers[2]).
		ReadIV(&x.Nonces[0]).ReadIV(&x.Nonces[1]).ReadIV(&x.Nonces[2])
	return
}
