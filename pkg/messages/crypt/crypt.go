package crypt

import (
	"crypto/cipher"
	
	"git-indra.lan/indra-labs/indra"
	"git-indra.lan/indra-labs/indra/pkg/crypto/ciph"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/cloak"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/messages/magicbytes"
	"git-indra.lan/indra-labs/indra/pkg/messages/reverse"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/splice"
	"git-indra.lan/indra-labs/indra/pkg/types"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

const (
	MagicString = "os"
	Len         = magicbytes.Len + nonce.IVLen + cloak.Len +
		pub.KeyLen
	ReverseLayerLen  = reverse.Len + Len
	ReverseHeaderLen = 3 * ReverseLayerLen
)

var (
	log               = log2.GetLogger(indra.PathBase)
	check             = log.E.Chk
	Magic             = slice.Bytes(MagicString)
	_     types.Onion = &Layer{}
)

// Layer message is the generic top level wrapper for a Layer. All
// following messages are wrapped inside this. This type provides the encryption
// for each crypt, and a header which a relay uses to determine what cipher to
// use. Everything in a message after this message is encrypted as specified.
type Layer struct {
	Depth                     int
	ToHeaderPub, ToPayloadPub *pub.Key
	From                      *prv.Key
	// The remainder here are for Decode.
	Nonce   nonce.IV
	Cloak   cloak.PubKey
	ToPriv  *prv.Key
	FromPub *pub.Key
	// The following field is only populated in the outermost crypt, and
	// passed in the `b slice.Bytes` parameter in both encode and decode,
	// this is created after first getting the Len of everything and
	// pre-allocating.
	slice.Bytes
	types.Onion
}

// func (x *Layer) String() string {
// 	return fmt.Sprintf("\n\tnonce: %x\n\tto: %x,\n\tfrom: %x,\n",
// 		x.Nonce, x.ToHeaderPub.ToBytes(), x.From.ToBytes())
// }

func (x *Layer) Insert(o types.Onion) { x.Onion = o }
func (x *Layer) Len() int {
	return Len + x.Onion.Len()
}
func (x *Layer) Encode(b slice.Bytes, c *slice.Cursor) {
	splice.Splice(b, c).
		Magic(Magic).
		IV(x.Nonce).
		Cloak(x.ToHeaderPub).
		Pubkey(pub.Derive(x.From))
	start := int(*c)
	// Call the tree of onions to perform their encoding.
	x.Onion.Encode(b, c)
	// Then we can encrypt the message segment
	var e error
	var blk cipher.Block
	if blk = ciph.GetBlock(x.From, x.ToHeaderPub); check(e) {
		panic(e)
	}
	end := len(b)
	switch {
	case x.Depth == 0:
	case x.Depth > 0:
		end = start + (x.Depth-1)*ReverseLayerLen
	default:
		panic("incorrect value for crypt sequence")
	}
	ciph.Encipher(blk, x.Nonce, b[start:end])
	if end != len(b) {
		if blk = ciph.GetBlock(x.From, x.ToPayloadPub); check(e) {
			panic(e)
		}
		ciph.Encipher(blk, x.Nonce, b[end:])
	}
}
func (x *Layer) Decode(b slice.Bytes, c *slice.Cursor) (e error) {
	if len(b[*c:]) < Len-magicbytes.Len {
		return magicbytes.TooShort(len(b[*c:]), Len-magicbytes.Len, "message")
	}
	splice.Splice(b, c).
		ReadIV(&x.Nonce).
		ReadCloak(&x.Cloak).
		ReadPubkey(&x.FromPub)
	// A further step is required which decrypts the remainder of the bytes
	// after finding the private key corresponding to the Cloak.
	return
}

// Decrypt requires the prv.Key to be located from the Cloak, using the FromPub
// key to derive the shared secret, and then decrypts the rest of the message.
func (x *Layer) Decrypt(prk *prv.Key, b slice.Bytes, c *slice.Cursor) {
	ciph.Encipher(ciph.GetBlock(prk, x.FromPub), x.Nonce, b[*c:])
}
