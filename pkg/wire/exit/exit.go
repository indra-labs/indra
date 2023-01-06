package exit

import (
	"github.com/indra-labs/indra"
	log2 "github.com/indra-labs/indra/pkg/log"
	"github.com/indra-labs/indra/pkg/nonce"
	"github.com/indra-labs/indra/pkg/sha256"
	"github.com/indra-labs/indra/pkg/slice"
	"github.com/indra-labs/indra/pkg/types"
	"github.com/indra-labs/indra/pkg/wire/magicbytes"
	"github.com/davecgh/go-spew/spew"
)

const (
	MagicString = "ex"
	Len         = magicbytes.Len + slice.Uint16Len + 3*sha256.Len +
		slice.Uint32Len + nonce.IVLen*3
)

var (
	log               = log2.GetLogger(indra.PathBase)
	check             = log.E.Chk
	Magic             = slice.Bytes(MagicString)
	_     types.Onion = &OnionSkin{}
)

// OnionSkin exit messages are the layer of a message after two Forward packets
// that provides an exit address and
type OnionSkin struct {
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
	// Bytes are the message to be passed to the exit service.
	slice.Bytes
	types.Onion
}

func (x *OnionSkin) String() string {
	return spew.Sdump(x.Port, x.Ciphers, x.Nonces, x.Bytes.ToBytes())
}

func (x *OnionSkin) Inner() types.Onion   { return x.Onion }
func (x *OnionSkin) Insert(o types.Onion) { x.Onion = o }
func (x *OnionSkin) Len() int {
	return Len + x.Bytes.Len() + x.Onion.Len()
}

func (x *OnionSkin) Encode(b slice.Bytes, c *slice.Cursor) {
	copy(b[*c:c.Inc(magicbytes.Len)], Magic)
	port := slice.NewUint16()
	slice.EncodeUint16(port, int(x.Port))
	copy(b[*c:c.Inc(slice.Uint16Len)], port)
	copy(b[*c:c.Inc(sha256.Len)], x.Ciphers[0][:])
	copy(b[*c:c.Inc(sha256.Len)], x.Ciphers[1][:])
	copy(b[*c:c.Inc(sha256.Len)], x.Ciphers[2][:])
	copy(b[*c:c.Inc(nonce.IVLen)], x.Nonces[0][:])
	copy(b[*c:c.Inc(nonce.IVLen)], x.Nonces[1][:])
	copy(b[*c:c.Inc(nonce.IVLen)], x.Nonces[2][:])
	bytesLen := slice.NewUint32()
	slice.EncodeUint32(bytesLen, len(x.Bytes))
	copy(b[*c:c.Inc(slice.Uint32Len)], bytesLen)
	copy(b[*c:c.Inc(len(x.Bytes))], x.Bytes)
	x.Onion.Encode(b, c)
}

func (x *OnionSkin) Decode(b slice.Bytes, c *slice.Cursor) (e error) {
	if len(b[*c:]) < Len-magicbytes.Len {
		return magicbytes.TooShort(len(b[*c:]), Len-magicbytes.Len, string(Magic))
	}
	x.Port = uint16(slice.DecodeUint16(b[*c:c.Inc(slice.Uint16Len)]))
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
	bytesLen := slice.DecodeUint32(b[*c:c.Inc(slice.Uint32Len)])
	x.Bytes = b[*c:c.Inc(bytesLen)]
	return
}
