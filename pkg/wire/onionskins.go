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
	ForwardMagic         = slice.Bytes("fwd")
	ExitMagic            = slice.Bytes("exi")
	ReturnMagic          = slice.Bytes("rtn")
	CipherMagic          = slice.Bytes("cif")
	PurchaseMagic        = slice.Bytes("prc")
	SessionMagic         = slice.Bytes("ses")
	AcknowledgementMagic = slice.Bytes("ack")
	ResponseMagic        = slice.Bytes("res")
	TokenMagic           = slice.Bytes("tok")
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
func (on *Message) Len() int       { return MagicLen + OnionHeaderLen + on.Onion.Len() }

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

// Forward is just an IP address and a wrapper for another message.
type Forward struct {
	net.IP
	Onion
}

var _ Onion = &Forward{}

func (fw *Forward) Inner() Onion   { return fw.Onion }
func (fw *Forward) Insert(o Onion) { fw.Onion = o }
func (fw *Forward) Len() int       { return MagicLen + len(fw.IP) + 1 + fw.Onion.Len() }

func (fw *Forward) Encode(o slice.Bytes, c *slice.Cursor) {
	copy(o[*c:c.Inc(MagicLen)], ForwardMagic)
	o[*c] = byte(len(fw.IP))
	copy(o[c.Inc(1):c.Inc(len(fw.IP))], fw.IP)
	fw.Onion.Encode(o, c)
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

// Return messages are distinct from Forward messages in that the header
// encryption uses a different secret than the payload. The magic bytes signal
// this to the relay that receives this, which then looks up the Return key
// matching the To address in the message header.
type Return struct {
	// IP is the address of the next relay in the return leg of a circuit.
	net.IP
	// The Key here should be the Return key matching the IP of the relay.
	// The header provided in a previous Exit message uses the Forward key
	// so that the Exit node cannot decrypt the header and discover the
	// return path.
	pub.Key
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

// Cipher delivers a public key to be used in association with a Return
// specifically in the situation of a node bootstrapping that doesn't have
// sessions yet. The ID allows the client to associate the Cipher to the
// Purchase.
type Cipher struct {
	nonce.ID
	Key pub.Bytes
	Onion
}

var _ Onion = &Cipher{}

func (ci *Cipher) Inner() Onion   { return ci.Onion }
func (ci *Cipher) Insert(o Onion) { ci.Onion = o }
func (ci *Cipher) Len() int       { return MagicLen + pub.KeyLen + ci.Onion.Len() }

func (ci *Cipher) Encode(o slice.Bytes, c *slice.Cursor) {
	copy(o[*c:c.Inc(nonce.IDLen)], ci.ID[:])
	copy(o[*c:c.Inc(MagicLen)], CipherMagic)
	copy(o[c.Inc(1):c.Inc(sha256.Len)], ci.Key[:])
	ci.Onion.Encode(o, c)
}

// Purchase is a message that is sent after first forwarding a Lighting payment
// of an amount corresponding to the number of bytes requested based on the
// price advertised for Exit traffic by a relay. The Receipt is the confirmation
// after requesting an Invoice for the amount and then paying it.
//
// This message contains a Return message, which enables payments to proxy
// forwards through two hops to the router that will issue the Session, plus two
// more Return layers for carrying the Session back to the client.
//
// Purchases have an ID created by the client.
type Purchase struct {
	Value uint64
	Onion
}

var _ Onion = &Purchase{}

func (pr *Purchase) Inner() Onion   { return pr.Onion }
func (pr *Purchase) Insert(o Onion) { pr.Onion = o }
func (pr *Purchase) Len() int {
	return MagicLen + slice.Uint64Len + pr.Onion.Len()
}

func (pr *Purchase) Encode(o slice.Bytes, c *slice.Cursor) {
	copy(o[*c:c.Inc(MagicLen)], PurchaseMagic)
	value := slice.NewUint64()
	slice.EncodeUint64(value, pr.Value)
	copy(o[*c:c.Inc(slice.Uint64Len)], value)
	pr.Onion.Encode(o, c)
}

// Session is a message containing two public keys which identify to a relay the
// session to account bytes on, this is wrapped in two Return message layers by
// the seller. Forward keys are used for encryption in Forward and Exit
// messages, and Return keys are separate and are only known to the client and
// relay that issues a Session, ensuring that the Exit cannot see the inner
// layers of the Return messages.
type Session struct {
	ForwardKey pub.Bytes
	ReturnKey  pub.Bytes
	Onion
}

var _ Onion = &Session{}

func (se *Session) Inner() Onion   { return se.Onion }
func (se *Session) Insert(o Onion) { se.Onion = o }
func (se *Session) Len() int {
	return MagicLen + pub.KeyLen*2 + se.Onion.Len()
}

func (se *Session) Encode(o slice.Bytes, c *slice.Cursor) {
	copy(o[*c:c.Inc(MagicLen)], SessionMagic)
	copy(o[*c:c.Inc(pub.KeyLen)], se.ForwardKey[:])
	copy(o[*c:c.Inc(pub.KeyLen)], se.ReturnKey[:])
	se.Onion.Encode(o, c)
}

// The remaining methods are terminals, all constructed Onion structures
// should have one of these as the last element otherwise the second last call
// to Encode will panic with a nil.

// Acknowledgement messages just contain a nonce ID, these are used to terminate
// ping and Cipher onion messages that confirm relaying was successful.
type Acknowledgement struct {
	nonce.ID
}

var _ Onion = &Acknowledgement{}

func (ak *Acknowledgement) Inner() Onion   { return nil }
func (ak *Acknowledgement) Insert(_ Onion) {}
func (ak *Acknowledgement) Len() int       { return MagicLen + nonce.IDLen }

func (ak *Acknowledgement) Encode(o slice.Bytes, c *slice.Cursor) {
	copy(o[*c:c.Inc(MagicLen)], AcknowledgementMagic)
	copy(o[*c:c.Inc(pub.KeyLen)], ak.ID[:])
}

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

// A Token is a 32 byte value. TODO: not sure we need this?
type Token sha256.Hash

var _ Onion = Token{}

func (tk Token) Inner() Onion   { return nil }
func (tk Token) Insert(_ Onion) {}
func (tk Token) Len() int       { return MagicLen + sha256.Len }

func (tk Token) Encode(o slice.Bytes, c *slice.Cursor) {
	copy(o[*c:c.Inc(MagicLen)], TokenMagic)
	copy(o[*c:c.Inc(sha256.Len)], tk[:])
}
