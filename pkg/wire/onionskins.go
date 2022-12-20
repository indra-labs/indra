package wire

import (
	"crypto/cipher"
	"net"

	"github.com/Indra-Labs/indra"
	"github.com/Indra-Labs/indra/pkg/ciph"
	"github.com/Indra-Labs/indra/pkg/key/address"
	"github.com/Indra-Labs/indra/pkg/key/prv"
	"github.com/Indra-Labs/indra/pkg/key/pub"
	"github.com/Indra-Labs/indra/pkg/nonce"
	"github.com/Indra-Labs/indra/pkg/sha256"
	"github.com/Indra-Labs/indra/pkg/slice"
	log2 "github.com/cybriq/proc/pkg/log"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

// MagicLen is 3 to make it infeasible that the wrong cipher will yield a
// valid Magic string as listed below.
const MagicLen = 3

var (
	ConfirmationMagic = slice.Bytes("cnf")
	ForwardMagic      = slice.Bytes("fwd")
	ExitMagic         = slice.Bytes("exi")
	ReturnMagic       = slice.Bytes("rtn")
	CipherMagic       = slice.Bytes("cif")
	PurchaseMagic     = slice.Bytes("prc")
	SessionMagic      = slice.Bytes("ses")
	ResponseMagic     = slice.Bytes("res")
	TokenMagic        = slice.Bytes("tok")
)

// Onion is an interface for the layers of messages each encrypted inside a
// Message, which provides the cipher for the inner layers inside it.
type Onion interface {
	Encode(o slice.Bytes, c *slice.Cursor)
	Len() int
	Inner() Onion
	Insert(on Onion)
}

// Message is the generic top level wrapper for an Onion. All following messages
// are wrapped inside this. This type provides the encryption for each layer,
// and a header which a relay uses to determine what cipher to use.
type Message struct {
	To   *address.Sender
	From *prv.Key
	// The following field is only populated in the outermost layer.
	slice.Bytes
	Onion
}

const OnionHeaderLen = 4 + nonce.IVLen + address.Len + pub.KeyLen

var _ Onion = &Message{}

func (on *Message) Inner() Onion   { return on.Onion }
func (on *Message) Insert(o Onion) { on.Onion = o }
func (on *Message) Len() int {
	return MagicLen + OnionHeaderLen + on.Onion.Len()
}

func (on *Message) Encode(o slice.Bytes, c *slice.Cursor) {
	// The first level message contains the Bytes, but the inner layers do
	// not. The inner layers will be passed this buffer, but the first needs
	// to have it copied from its original location.
	if o == nil {
		o = on.Bytes
	}
	// We write the checksum last so save the cursor position here.
	checkStart := *c
	checkEnd := checkStart + 4
	// Generate a new nonce and copy it in.
	n := nonce.New()
	copy(o[c.Inc(4):c.Inc(nonce.IVLen)], n[:])
	// Derive the cloaked key and copy it in.
	to := on.To.GetCloak()
	copy(o[*c:c.Inc(address.Len)], to[:])
	// Call the tree of onions to perform their encoding.
	on.Onion.Encode(o, c)
	// Then we can encrypt the message segment
	var e error
	var blk cipher.Block
	if blk = ciph.GetBlock(on.From, on.To.Key); check(e) {
		panic(e)
	}
	ciph.Encipher(blk, n, o[checkEnd:])
	// Get the hash of the message and truncate it to the checksum at the
	// start of the message. Every layer of the onion has a Message and an
	// onion inside it, the Message takes care of the encryption. This saves
	// on complications as every layer is header first, message after, with
	// wrapped messages inside each message afterwards.
	hash := sha256.Single(o[checkEnd:])
	copy(o[checkStart:checkEnd], hash[:4])
}

// Confirmation is an encryption layer for messages returned to the client on
// the inside of an onion, for Ping and Cipher messages, providing a
// confirmation of the transit of the onion through its encoded route.
//
// It is encrypted because otherwise internal identifiers could be leaked and
// potentially reveal something about the entropy of a client/relay.
//
// In order to speed up recognition, the key of the table of pending Ping and
// Cipher messages will include the last hop that will deliver this layer of the
// onion - there can be more than one up in the air at a time, but they are
// randomly selected, so they will generally be a much smaller subset versus the
// current full set of Session s currently open.
type Confirmation struct {
	nonce.ID
}

var _ Onion = &Confirmation{}

func (cf *Confirmation) Inner() Onion   { return nil }
func (cf *Confirmation) Insert(o Onion) {}
func (cf *Confirmation) Len() int {
	return MagicLen + nonce.IDLen
}

func (cf *Confirmation) Encode(o slice.Bytes, c *slice.Cursor) {
	copy(o[*c:c.Inc(MagicLen)], ConfirmationMagic)
	// Copy in the ID.
	copy(o[*c:c.Inc(nonce.IDLen)], cf.ID[:])
}

// Forward is just an IP address and a wrapper for another message.
type Forward struct {
	net.IP
	Onion
}

var _ Onion = &Forward{}

func (fw *Forward) Inner() Onion   { return fw.Onion }
func (fw *Forward) Insert(o Onion) { fw.Onion = o }
func (fw *Forward) Len() int {
	return MagicLen + len(fw.IP) + 1 + fw.Onion.Len()
}

func (fw *Forward) Encode(o slice.Bytes, c *slice.Cursor) {
	copy(o[*c:c.Inc(MagicLen)], ForwardMagic)
	o[*c] = byte(len(fw.IP))
	copy(o[c.Inc(1):c.Inc(len(fw.IP))], fw.IP)
	fw.Onion.Encode(o, c)
}

// Return messages are distinct from Forward messages in that the header
// encryption uses a different secret than the payload. The magic bytes signal
// this to the relay that receives this, which then looks up the Return key
// matching the To address in the message header.
type Return struct {
	// IP is the address of the next relay in the return leg of a circuit.
	net.IP
	Onion
}

var _ Onion = &Return{}

func (rt *Return) Inner() Onion   { return rt.Onion }
func (rt *Return) Insert(o Onion) { rt.Onion = o }
func (rt *Return) Len() int {
	return MagicLen + len(rt.IP) + 1 + rt.Onion.Len()
}

func (rt *Return) Encode(o slice.Bytes, c *slice.Cursor) {
	copy(o[*c:c.Inc(MagicLen)], ReturnMagic)
	o[*c] = byte(len(rt.IP))
	copy(o[c.Inc(1):c.Inc(len(rt.IP))], rt.IP)
	rt.Onion.Encode(o, c)
}

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
	Onion
}

var _ Onion = &Exit{}

func (ex *Exit) Inner() Onion   { return ex.Onion }
func (ex *Exit) Insert(o Onion) { ex.Onion = o }
func (ex *Exit) Len() int {
	return MagicLen + slice.Uint16Len + 3*sha256.Len + ex.Bytes.Len() +
		ex.Onion.Len()
}

func (ex *Exit) Encode(o slice.Bytes, c *slice.Cursor) {
	copy(o[*c:c.Inc(MagicLen)], ExitMagic)
	port := slice.NewUint16()
	slice.EncodeUint16(port, int(ex.Port))
	copy(o[*c:c.Inc(slice.Uint16Len)], port)
	copy(o[*c:c.Inc(sha256.Len)], ex.Cipher[0][:])
	copy(o[*c:c.Inc(sha256.Len)], ex.Cipher[1][:])
	copy(o[*c:c.Inc(sha256.Len)], ex.Cipher[1][:])
	bytesLen := slice.NewUint32()
	slice.EncodeUint32(bytesLen, len(ex.Bytes))
	copy(o[*c:c.Inc(slice.Uint32Len)], bytesLen)
	copy(o[*c:c.Inc(len(ex.Bytes))], ex.Bytes)
	ex.Onion.Encode(o, c)

}

// Cipher delivers a public key to be used in association with a Return
// specifically in the situation of a node bootstrapping that doesn't have
// sessions yet. The Forward key will appear in the pre-formed header, but the
// cipher provided to the seller will correspond with
type Cipher struct {
	Header, Payload *prv.Key
	Onion
}

var _ Onion = &Cipher{}

func (ci *Cipher) Inner() Onion   { return ci.Onion }
func (ci *Cipher) Insert(o Onion) { ci.Onion = o }
func (ci *Cipher) Len() int {
	return MagicLen + pub.KeyLen + ci.Onion.Len()
}

func (ci *Cipher) Encode(o slice.Bytes, c *slice.Cursor) {
	copy(o[*c:c.Inc(MagicLen)], CipherMagic)
	hdr, pld := ci.Header.ToBytes(), ci.Payload.ToBytes()
	copy(o[c.Inc(1):c.Inc(prv.KeyLen)], hdr[:])
	copy(o[c.Inc(1):c.Inc(prv.KeyLen)], pld[:])
	ci.Onion.Encode(o, c)
}

// Purchase is a message that is sent after first forwarding a Lighting payment
// of an amount corresponding to the number of bytes requested based on the
// price advertised for Exit traffic by a relay.
//
// The Return bytes contain the message header that is prepended to a Session
// message which contains the pair of keys associated with the Session that is
// purchased.
type Purchase struct {
	Value uint64
	// Ciphers is a set of 3 symmetric ciphers that are to be used in their
	// given order over the reply message from the service.
	Ciphers [3]sha256.Hash
	// Return is the pre-formed header which uses different private keys to
	// the ones used to create the Ciphers above, meaning the seller can
	// encrypt the payload to be correctly decrypted by the Return hops, but
	// cannot decrypt the header, which would reveal the return path.
	Return slice.Bytes
	Onion
}

var _ Onion = &Purchase{}

func (pr *Purchase) Inner() Onion   { return pr.Onion }
func (pr *Purchase) Insert(o Onion) { pr.Onion = o }
func (pr *Purchase) Len() int {
	return MagicLen + slice.Uint64Len + len(pr.Return) + pr.Onion.Len()
}

func (pr *Purchase) Encode(o slice.Bytes, c *slice.Cursor) {
	copy(o[*c:c.Inc(MagicLen)], PurchaseMagic)
	value := slice.NewUint64()
	slice.EncodeUint64(value, pr.Value)
	copy(o[*c:c.Inc(slice.Uint64Len)], value)
	copy(o[*c:c.Inc(sha256.Len)], pr.Ciphers[0][:])
	copy(o[*c:c.Inc(sha256.Len)], pr.Ciphers[1][:])
	copy(o[*c:c.Inc(sha256.Len)], pr.Ciphers[1][:])
	copy(o[*c:c.Inc(len(pr.Return))], pr.Return)
	pr.Onion.Encode(o, c)
}

// Session is a message containing two public keys which identify to a relay how
// to decrypt the header in a Return message, using the HeaderKey, and the
// payload, which uses the PayloadKey. There is two keys in order to prevent the
// Exit node from being able to decrypt the header, but enable it to encrypt the
// payload, and Return relay hops have these key pairs and identify the
// HeaderKey and then know they can unwrap their layer of the payload using the
// PayloadKey.
//
// Clients use the HeaderKey, cloaked, in their messages for the seller relay,
// in the header, and use the PayloadKey as the public key half with ECDH and
// their generated private key which produces the public key that is placed in
// the header.
type Session struct {
	HeaderKey  *pub.Key
	PayloadKey *pub.Key
	Onion
}

var _ Onion = &Session{}

func (se *Session) Inner() Onion   { return se.Onion }
func (se *Session) Insert(o Onion) { se.Onion = o }
func (se *Session) Len() int {
	return MagicLen + pub.KeyLen*2 + se.Onion.Len()
}

func (se *Session) Encode(o slice.Bytes, c *slice.Cursor) {
	hdr, pld := se.HeaderKey.ToBytes(), se.PayloadKey.ToBytes()
	copy(o[*c:c.Inc(MagicLen)], SessionMagic)
	copy(o[*c:c.Inc(pub.KeyLen)], hdr[:])
	copy(o[*c:c.Inc(pub.KeyLen)], pld[:])
	se.Onion.Encode(o, c)
}

// The remaining types are terminals, all constructed Onion structures
// should have one of these as the last element otherwise the second last call
// to Encode will panic with a nil.

// Response messages are what are carried back via Return messages from an Exit.
type Response slice.Bytes

var _ Onion = Response{}

func (rs Response) Inner() Onion   { return nil }
func (rs Response) Insert(_ Onion) {}
func (rs Response) Len() int       { return MagicLen + len(rs) + 4 }

func (rs Response) Encode(o slice.Bytes, c *slice.Cursor) {
	copy(o[*c:c.Inc(MagicLen)], ResponseMagic)
	bytesLen := slice.NewUint32()
	slice.EncodeUint32(bytesLen, len(rs))
	copy(o[*c:c.Inc(slice.Uint32Len)], bytesLen)
	copy(o[*c:c.Inc(len(rs))], rs)
}

// A Token is a 32 byte value.
type Token sha256.Hash

var _ Onion = Token{}

func (tk Token) Inner() Onion   { return nil }
func (tk Token) Insert(_ Onion) {}
func (tk Token) Len() int       { return MagicLen + sha256.Len }

func (tk Token) Encode(o slice.Bytes, c *slice.Cursor) {
	copy(o[*c:c.Inc(MagicLen)], TokenMagic)
	copy(o[*c:c.Inc(sha256.Len)], tk[:])
}
