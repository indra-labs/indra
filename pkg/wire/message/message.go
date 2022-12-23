package message

import (
	"crypto/cipher"

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
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

// Type message is the generic top level wrapper for an Onion. All following messages
// are wrapped inside this. This type provides the encryption for each layer,
// and a header which a relay uses to determine what cipher to use.
type Type struct {
	To   *address.Sender
	From *prv.Key
	// The following field is only populated in the outermost layer.
	slice.Bytes
	types.Onion
}

const MinLen = magicbytes.Len + slice.Uint32Len + nonce.IVLen +
	address.Len + pub.KeyLen

var _ types.Onion = &Type{}

func (x *Type) Inner() types.Onion   { return x.Onion }
func (x *Type) Insert(o types.Onion) { x.Onion = o }
func (x *Type) Len() int {
	return MinLen + x.Onion.Len()
}

func (x *Type) Encode(o slice.Bytes, c *slice.Cursor) {
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

func (x *Type) Decode(b slice.Bytes) (e error) {
	minLen := MinLen
	if len(b) < minLen {
		return magicbytes.TooShort(len(b), minLen, "message")
	}
	sc := slice.Cursor(0)
	c := &sc
	_ = c

	return
}
