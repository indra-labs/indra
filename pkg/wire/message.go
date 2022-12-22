package wire

import (
	"crypto/cipher"

	"github.com/Indra-Labs/indra/pkg/ciph"
	"github.com/Indra-Labs/indra/pkg/key/address"
	"github.com/Indra-Labs/indra/pkg/key/prv"
	"github.com/Indra-Labs/indra/pkg/key/pub"
	"github.com/Indra-Labs/indra/pkg/nonce"
	"github.com/Indra-Labs/indra/pkg/sha256"
	"github.com/Indra-Labs/indra/pkg/slice"
	"github.com/Indra-Labs/indra/pkg/types"
)

// Message is the generic top level wrapper for an Onion. All following messages
// are wrapped inside this. This type provides the encryption for each layer,
// and a header which a relay uses to determine what cipher to use.
type Message struct {
	To   *address.Sender
	From *prv.Key
	// The following field is only populated in the outermost layer.
	slice.Bytes
	types.Onion
}

const OnionHeaderLen = 4 + nonce.IVLen + address.Len + pub.KeyLen

var _ types.Onion = &Message{}

func (x *Message) Inner() types.Onion   { return x.Onion }
func (x *Message) Insert(o types.Onion) { x.Onion = o }
func (x *Message) Len() int {
	return MagicLen + OnionHeaderLen + x.Onion.Len()
}

func (x *Message) Encode(o slice.Bytes, c *slice.Cursor) {
	// The first level message contains the Bytes, but the inner layers do
	// not. The inner layers will be passed this buffer, but the first needs
	// to have it copied from its original location.
	if o == nil {
		o = x.Bytes
	}
	// We write the checksum last so save the cursor position here.
	checkStart := *c
	checkEnd := checkStart + 4
	// Generate a new nonce and copy it in.
	n := nonce.New()
	copy(o[c.Inc(4):c.Inc(nonce.IVLen)], n[:])
	// Derive the cloaked key and copy it in.
	to := x.To.GetCloak()
	copy(o[*c:c.Inc(address.Len)], to[:])
	// Derive the public key from the From key and copy in.
	pubKey := pub.Derive(x.From).ToBytes()
	copy(o[*c:c.Inc(pub.KeyLen)], pubKey[:])
	// Call the tree of onions to perform their encoding.
	x.Onion.Encode(o, c)

	// Then we can encrypt the message segment
	var e error
	var blk cipher.Block
	if blk = ciph.GetBlock(x.From, x.To.Key); check(e) {
		panic(e)
	}
	ciph.Encipher(blk, n, o[checkEnd:])
	// Get the hash of the message and truncate it to the checksum at the
	// start of the message. Every layer of the onion has a Header and an
	// onion inside it, the Header takes care of the encryption. This saves
	// x complications as every layer is header first, message after, with
	// wrapped messages inside each message afterwards.
	hash := sha256.Single(o[checkEnd:])
	copy(o[checkStart:checkEnd], hash[:4])
}

func (x *Message) Decode(b slice.Bytes, c *slice.Cursor) (in interface{},
	e error) {

	return
}
