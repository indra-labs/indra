package exit

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
	MagicString = "ex"
	Len         = magicbytes.Len + slice.Uint16Len + 3*sha256.Len +
		slice.Uint32Len + nonce.IVLen*3 + nonce.IDLen
)

var (
	log               = log2.GetLogger(indra.PathBase)
	check             = log.E.Chk
	Magic             = slice.Bytes(MagicString)
	_     types.Onion = &Layer{}
)

// Layer exit messages are the crypt of a message after two Forward packets
// that provides an exit address and
type Layer struct {
	// Port identifies the type of service as well as being the port used by
	// the service to be relayed to. Notice there is no IP address, this is
	// because Indranet only forwards to exits of decentralised services
	// also running on the same machine. This service could be a proxy, of
	// course, if configured this way. This could be done by tunneling from
	// a local Socks5 proxy into Indranet and the exit node also having
	// this.
	Port uint16
	// Ciphers is a set of 3 symmetric ciphers that are to be used in their
	// given order over the reply message from the service.
	Ciphers [3]sha256.Hash
	// Nonces are the nonces to use with the cipher when creating the
	// encryption for the reply message.
	Nonces [3]nonce.IV
	nonce.ID
	// Bytes are the message to be passed to the exit service.
	slice.Bytes
	types.Onion
}

// func (x *Layer) String() string {
// 	return spew.Sdump(x.Port, x.Ciphers, x.Nonces, x.Bytes.ToBytes())
// }

func (x *Layer) Insert(o types.Onion) { x.Onion = o }
func (x *Layer) Len() int {
	return Len + x.Bytes.Len() + x.Onion.Len()
}

func (x *Layer) Encode(b slice.Bytes, c *slice.Cursor) {
	splice.Splice(b, c).
		Magic(Magic).
		Uint16(x.Port).
		Hash(x.Ciphers[0]).Hash(x.Ciphers[1]).Hash(x.Ciphers[2]).
		IV(x.Nonces[0]).IV(x.Nonces[1]).IV(x.Nonces[2]).
		ID(x.ID).
		Bytes(x.Bytes)
	x.Onion.Encode(b, c)
}

func (x *Layer) Decode(b slice.Bytes, c *slice.Cursor) (e error) {
	if len(b[*c:]) < Len-magicbytes.Len {
		return magicbytes.TooShort(len(b[*c:]), Len-magicbytes.Len, string(Magic))
	}
	splice.Splice(b, c).
		ReadUint16(&x.Port).
		ReadHash(&x.Ciphers[0]).ReadHash(&x.Ciphers[1]).ReadHash(&x.Ciphers[2]).
		ReadIV(&x.Nonces[0]).ReadIV(&x.Nonces[1]).ReadIV(&x.Nonces[2]).
		ReadID(&x.ID).
		ReadBytes(&x.Bytes)
	return
}
