package wire

import (
	"github.com/Indra-Labs/indra/pkg/sha256"
	"github.com/Indra-Labs/indra/pkg/slice"
	"github.com/Indra-Labs/indra/pkg/types"
)

// Exit messages are the layer of a message after two Forward packets that
// provides an exit address and
type Exit struct {
	// Port identifies the type of service as well as being the port used by
	// the service to be relayed to. Notice there is no IP address, this is
	// because Indranet only forwards to exits of decentralised services
	// also running on the same machine. This service could be a proxy, of
	// course, if configured this way. This could be done by tunneling from
	// a local Socks5 proxy into Indranet and the exit node also having
	// this.
	Port uint16
	// Cipher is a set of 3 symmetric ciphers that are to be used in their
	// given order over the reply message from the service.
	Cipher [3]sha256.Hash
	// Bytes are the message to be passed to the exit service.
	slice.Bytes
	types.Onion
}

var (
	ExitMagic             = slice.Bytes("exi")
	_         types.Onion = &Exit{}
)

func (x *Exit) Inner() types.Onion   { return x.Onion }
func (x *Exit) Insert(o types.Onion) { x.Onion = o }
func (x *Exit) Len() int {
	return MagicLen + slice.Uint16Len + 3*sha256.Len + x.Bytes.Len() +
		x.Onion.Len()
}

func (x *Exit) Encode(o slice.Bytes, c *slice.Cursor) {
	copy(o[*c:c.Inc(MagicLen)], ExitMagic)
	port := slice.NewUint16()
	slice.EncodeUint16(port, int(x.Port))
	copy(o[*c:c.Inc(slice.Uint16Len)], port)
	copy(o[*c:c.Inc(sha256.Len)], x.Cipher[0][:])
	copy(o[*c:c.Inc(sha256.Len)], x.Cipher[1][:])
	copy(o[*c:c.Inc(sha256.Len)], x.Cipher[1][:])
	bytesLen := slice.NewUint32()
	slice.EncodeUint32(bytesLen, len(x.Bytes))
	copy(o[*c:c.Inc(slice.Uint32Len)], bytesLen)
	copy(o[*c:c.Inc(len(x.Bytes))], x.Bytes)
	x.Onion.Encode(o, c)

}

func (x *Exit) Decode(b slice.Bytes, c *slice.Cursor) (in interface{},
	e error) {

	magic := ExitMagic
	if !CheckMagic(b, magic) {
		return ReturnError(ErrWrongMagic, x, b, magic)
	}

	return
}
