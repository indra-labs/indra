package message

import (
	"crypto/cipher"
	"fmt"

	"github.com/Indra-Labs/indra"
	"github.com/Indra-Labs/indra/pkg/ciph"
	"github.com/Indra-Labs/indra/pkg/key/address"
	"github.com/Indra-Labs/indra/pkg/key/prv"
	"github.com/Indra-Labs/indra/pkg/key/pub"
	"github.com/Indra-Labs/indra/pkg/nonce"
	"github.com/Indra-Labs/indra/pkg/sha256"
	"github.com/Indra-Labs/indra/pkg/slice"
	"github.com/Indra-Labs/indra/pkg/types"
	"github.com/Indra-Labs/indra/pkg/wire/magicbytes"
	log2 "github.com/cybriq/proc/pkg/log"
)

var (
	log                     = log2.GetLogger(indra.PathBase)
	check                   = log.E.Chk
	MagicString             = "mg"
	Magic                   = slice.Bytes(MagicString)
	_           types.Onion = &OnionSkin{}
)

// OnionSkin message is the generic top level wrapper for an OnionSkin. All following
// messages are wrapped inside this. This type provides the encryption for each
// layer, and a header which a relay uses to determine what cipher to use.
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

const MinLen = magicbytes.Len + nonce.IVLen +
	address.Len + pub.KeyLen + slice.Uint32Len

func (x *OnionSkin) Inner() types.Onion   { return x.Onion }
func (x *OnionSkin) Insert(o types.Onion) { x.Onion = o }
func (x *OnionSkin) Len() int {
	return MinLen + x.Onion.Len()
}

func (x *OnionSkin) Encode(b slice.Bytes, c *slice.Cursor) {
	// The first level message contains the Bytes, but the inner layers do
	// not. The inner layers will be passed this buffer, but the first needs
	// to have it copied from its original location.
	if b == nil {
		b = x.Bytes
	}
	// Magic after the check, so it is part of the checksum.
	copy(b[*c:c.Inc(magicbytes.Len)], Magic)
	// Generate a new nonce and copy it in.
	n := nonce.New()
	copy(b[c.Inc(4):c.Inc(nonce.IVLen)], n[:])
	// Derive the cloaked key and copy it in.
	to := x.To.GetCloak()
	copy(b[*c:c.Inc(address.Len)], to[:])
	// Derive the public key from the From key and copy in.
	pubKey := pub.Derive(x.From).ToBytes()
	copy(b[*c:c.Inc(pub.KeyLen)], pubKey[:])
	// Encode the remaining data size of the message below. This will also
	// be the entire remaining part of the message buffer.
	mLen := len(b[*c:]) - slice.Uint32Len
	length := slice.NewUint32()
	slice.EncodeUint32(length, mLen)
	copy(b[*c:c.Inc(mLen)], b[*c:])
	start := *c
	// Call the tree of onions to perform their encoding.
	x.Onion.Encode(b, c)
	// Then we can encrypt the message segment
	var e error
	var blk cipher.Block
	if blk = ciph.GetBlock(x.From, x.To.Key); check(e) {
		panic(e)
	}
	ciph.Encipher(blk, n, b[start+MinLen:])
}

// Decode decodes a received message. Note that it only gets the relevant data
// from the header, a subsequent process must be performed to find the prv.Key
// corresponding to the Cloak and the pub.Key together forming the cipher secret
// needed to decrypt the remainder of the bytes.
func (x *OnionSkin) Decode(b slice.Bytes, c *slice.Cursor) (e error) {
	if len(b[*c:]) < MinLen {
		return magicbytes.TooShort(len(b[*c:]), MinLen, "message")
	}
	chek := b[*c:c.Inc(4)]
	start := int(*c)
	var n nonce.IV
	copy(n[:], b[c.Inc(magicbytes.Len):c.Inc(nonce.IVLen)])
	copy(x.Nonce[:], n[:])
	copy(x.Cloak[:], b[*c:c.Inc(address.Len)])
	if x.FromPub, e = pub.FromBytes(b[*c:c.Inc(pub.KeyLen)]); check(e) {
		return
	}
	length := slice.DecodeUint32(b[*c:c.Inc(slice.Uint32Len)])
	if length < len(b[*c:]) {
		e = fmt.Errorf("not enough remaining bytes as specified in"+
			" length field, got: %d expected %d",
			length, len(b[*c:]))
	}
	hash := sha256.Single(b[start : start+length])
	if string(hash[:4]) != string(chek) {
		return fmt.Errorf("message decode fail checksum")
	}
	// Snip out bytes for this layer from the remainder, until the length
	// indicated by the length prefix. Cursor will now be at the beginning
	// of the next layer's messages.
	x.Bytes = b[*c:c.Inc(length)]
	// A further step is required which decrypts the remainder of the bytes
	// after finding the private key corresponding to the Cloak.
	return
}

// Decrypt requires the prv.Key to be located from the Cloak, using the
// FromPub key to derive the shared secret, and then decrypts the rest of the
// message.
func (x *OnionSkin) Decrypt(prk *prv.Key) {
	ciph.Encipher(ciph.GetBlock(prk, x.FromPub), x.Nonce, x.Bytes)
}
