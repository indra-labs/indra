package layer

import (
	"crypto/cipher"
	"fmt"

	"github.com/indra-labs/indra"
	"github.com/indra-labs/indra/pkg/ciph"
	"github.com/indra-labs/indra/pkg/key/cloak"
	"github.com/indra-labs/indra/pkg/key/prv"
	"github.com/indra-labs/indra/pkg/key/pub"
	log2 "github.com/indra-labs/indra/pkg/log"
	"github.com/indra-labs/indra/pkg/nonce"
	"github.com/indra-labs/indra/pkg/slice"
	"github.com/indra-labs/indra/pkg/types"
	"github.com/indra-labs/indra/pkg/wire/magicbytes"
)

const (
	MagicString = "os"
	Len         = magicbytes.Len + nonce.IVLen + cloak.Len + pub.KeyLen
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
	To   *pub.Key
	From *prv.Key
	// The remainder here are for Decode.
	Nonce   nonce.IV
	Cloak   cloak.PubKey
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
	to := cloak.GetCloak(x.To)
	copy(b[*c:c.Inc(cloak.Len)], to[:])
	// Derive the public key from the From key and copy in.
	pubKey := pub.Derive(x.From).ToBytes()
	copy(b[*c:c.Inc(pub.KeyLen)], pubKey[:])
	start := *c
	// Call the tree of onions to perform their encoding.
	x.Onion.Encode(b, c)
	// Then we can encrypt the message segment
	var e error
	var blk cipher.Block
	if blk = ciph.GetBlock(x.From, x.To); check(e) {
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
	copy(x.Cloak[:], b[*c:c.Inc(cloak.Len)])
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
