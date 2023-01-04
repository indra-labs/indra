package layer

import (
	"crypto/cipher"
	"fmt"

	"github.com/Indra-Labs/indra"
	"github.com/Indra-Labs/indra/pkg/ciph"
	"github.com/Indra-Labs/indra/pkg/key/address"
	"github.com/Indra-Labs/indra/pkg/key/prv"
	"github.com/Indra-Labs/indra/pkg/key/pub"
	"github.com/Indra-Labs/indra/pkg/nonce"
	"github.com/Indra-Labs/indra/pkg/slice"
	"github.com/Indra-Labs/indra/pkg/types"
	"github.com/Indra-Labs/indra/pkg/wire/magicbytes"
	log2 "github.com/cybriq/proc/pkg/log"
)

const (
	MagicString = "os"
	Len         = magicbytes.Len + nonce.IVLen + address.Len + pub.KeyLen
)

var (
	log               = log2.GetLogger(indra.PathBase)
	check             = log.E.Chk
	Magic             = slice.Bytes(MagicString)
	_     types.Onion = &OnionSkin{}
)

// OnionSkin message is the generic top level wrapper for an OnionSkin. All
// following messages are wrapped inside this. This type provides the encryption
// for each layer, and a header which a relay uses to determine what cipher to
// use.
type OnionSkin struct {
	To   *address.Sender
	From *prv.Key
	// The remainder here are for Decode.
	Nonce   nonce.IV
	Cloak   address.Cloaked
	ToPriv  *prv.Key
	FromPub *pub.Key
	// The following field is only populated in the outermost layer, and
	// passed in the `b slice.Bytes` parameter in both encode and decode,
	// this is created after first getting the Len of everything and
	// pre-allocating.
	slice.Bytes
	types.Onion
}

func (x *OnionSkin) String() string {
	return fmt.Sprintf("\n\tnonce: %x\n\tto: %x,\n\tfrom: %x,\n",
		x.Nonce, x.To.ToBytes(), x.From.ToBytes())
}

func (x *OnionSkin) Inner() types.Onion   { return x.Onion }
func (x *OnionSkin) Insert(o types.Onion) { x.Onion = o }
func (x *OnionSkin) Len() int {
	return Len + x.Onion.Len()
}

func (x *OnionSkin) Encode(b slice.Bytes, c *slice.Cursor) {
	copy(b[*c:c.Inc(magicbytes.Len)], Magic)
	copy(b[*c:c.Inc(nonce.IVLen)], x.Nonce[:])
	// Derive the cloaked key and copy it in.
	to := x.To.GetCloak()
	copy(b[*c:c.Inc(address.Len)], to[:])
	// Derive the public key from the From key and copy in.
	pubKey := pub.Derive(x.From).ToBytes()
	copy(b[*c:c.Inc(pub.KeyLen)], pubKey[:])
	start := *c
	// Call the tree of onions to perform their encoding.
	x.Onion.Encode(b, c)
	// Then we can encrypt the message segment
	var e error
	var blk cipher.Block
	if blk = ciph.GetBlock(x.From, x.To.Key); check(e) {
		panic(e)
	}
	ciph.Encipher(blk, x.Nonce, b[start:])
}

// Decode decodes a received OnionSkin. The entire remainder of the message is
// encrypted by this layer.
func (x *OnionSkin) Decode(b slice.Bytes, c *slice.Cursor) (e error) {
	if len(b[*c:]) < Len-magicbytes.Len {
		return magicbytes.TooShort(len(b[*c:]), Len-magicbytes.Len, "message")
	}
	copy(x.Nonce[:], b[*c:c.Inc(nonce.IVLen)])
	copy(x.Cloak[:], b[*c:c.Inc(address.Len)])
	if x.FromPub, e = pub.FromBytes(b[*c:c.Inc(pub.KeyLen)]); check(e) {
		return
	}
	// A further step is required which decrypts the remainder of the bytes
	// after finding the private key corresponding to the Cloak.
	return
}

// Decrypt requires the prv.Key to be located from the Cloak, using the FromPub
// key to derive the shared secret, and then decrypts the rest of the message.
func (x *OnionSkin) Decrypt(prk *prv.Key, b slice.Bytes, c *slice.Cursor) {
	ciph.Encipher(ciph.GetBlock(prk, x.FromPub), x.Nonce, b[*c:])
}
