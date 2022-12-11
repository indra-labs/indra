package wire

import (
	"net"

	"github.com/Indra-Labs/indra/pkg/ifc"
	"github.com/Indra-Labs/indra/pkg/key/pub"
	"github.com/Indra-Labs/indra/pkg/message"
	"github.com/Indra-Labs/indra/pkg/nonce"
)

// MagicLen is 3 to make it nearly impossible that the wrong cipher will yield a
// valid Magic string as listed below.
const MagicLen = 3

type MessageMagic string

type Message interface {
	Encode() (o ifc.Message)
}

var (
	ForwardMagic  = MessageMagic("fwd")
	ExitMagic     = MessageMagic("exi")
	ReturnMagic   = MessageMagic("rtn")
	CipherMagic   = MessageMagic("cif")
	PurchaseMagic = MessageMagic("prc")
	SessionMagic  = MessageMagic("ses")
)

// Forward is just an IP address and a wrapper for another message.
type Forward struct {
	*message.Addresses
	net.IP
	Message
}

func (fw *Forward) Encode() (o ifc.Message) {
	ipLen := len(fw.IP)
	msg := fw.Message.Encode()
	msg, _ = message.Encode(fw.To, fw.From, msg)
	fwd := make([]byte, ipLen+1)
	fwd[0] = byte(ipLen)
	copy(fwd[1:], fw.IP)
	o = append(append(ifc.Message(ForwardMagic), fwd...), msg...)
	return

}

func DecodeForward(msg ifc.Message) (rm *Forward, e error) {
	return
}

// Exit messages are the layer of a message after two Forward packets that
// provides an exit address and
type Exit struct {
	*message.Addresses
	// Port identifies the type of service as well as being the port used
	// by the service to be relayed to. Notice there is no IP address, this
	// is because Indranet only forwards to exits of decentralised services
	// also running on the same machine. This service could be a proxy, of
	// course, if configured this way.
	Port uint16
	// Cipher is a set of 3 symmetric ciphers that are to be used in their
	// given order over the reply message from the service.
	Cipher [3][32]byte
	// Return is the encoded message with the three hops using the Return
	// keys for the relevant relays, encrypted progressively. Note that this
	// message uses a different Cipher than the one above
	Return ifc.Message
	Message
}

func (ex *Exit) Encode() (o ifc.Message) {
	msg := ex.Message.Encode()
	msg, _ = message.Encode(ex.To, ex.From, msg)
	return
}

func DecodeExit(msg ifc.Message) (rm *Exit, e error) {
	return
}

// Return messages are distinct from Forward messages in that the header
// encryption uses a different secret than the payload. Relays identify this by
// encrypting the first 16 bytes of the header and if one of the magic bytes is
// not found
type Return struct {
	*message.Addresses
	nonce.ID
	// IP is the address of the next relay in the return leg of a circuit.
	net.IP
	Message
}

func (rt *Return) Encode() (o ifc.Message) {
	msg := rt.Message.Encode()
	msg, _ = message.Encode(rt.To, rt.From, msg)
	return
}

func DecodeReturn(msg ifc.Message) (rm *Return, e error) {
	return
}

// Cipher delivers a public key to be used in association with a Return with the
// matching ID. This is wrapped in two Forward messages and contains two layers
// of Forward messages, ensuring that the node cannot discover where this comes
// from but allows a Return public key to be provided for a Purchase.
type Cipher struct {
	*message.Addresses
	nonce.ID
	Key pub.Bytes
	Forward
}

func (ci *Cipher) Encode() (o ifc.Message) {
	msg := ci.Forward.Encode()
	msg, _ = message.Encode(ci.To, ci.From, msg)
	return
}

func DecodeCipher(msg ifc.Message) (rm *Cipher, e error) {
	return
}

// Purchase is a message that is sent after first forwarding a Lighting payment
// of an amount corresponding to the number of bytes requested based on the
// price advertised for Exit traffic by a relay. The Receipt is the confirmation
// after requesting an Invoice for the amount and then paying it.
//
// This message contains a Return message, which enables payments to proxy
// forwards through two hops to the router that will issue the Session, plus two
// more Return layers for carrying the Session back to the client.
type Purchase struct {
	*message.Addresses
	Bytes   int
	Receipt ifc.Message
	Return
}

func (pr *Purchase) Encode() (o ifc.Message) {
	msg := pr.Return.Encode()
	msg, _ = message.Encode(pr.To, pr.From, msg)
	return
}

func DecodePurchase(msg ifc.Message) (rm *Purchase, e error) {
	return
}

// Session is a message containing two public keys which identify to a relay the
// session to account bytes on, this is wrapped in two Return message layers by
// the seller. Forward keys are used for encryption in Forward and Exit
// messages, and Return keys are separate and are only known to the client and
// relay that issues a Session, ensuring that the Exit cannot see the inner
// layers of the Return messages.
type Session struct {
	*message.Addresses
	ForwardKey pub.Bytes
	ReturnKey  pub.Bytes
	Return
}

func (se *Session) Encode() (o ifc.Message) {
	msg := se.Return.Encode()
	msg, _ = message.Encode(se.To, se.From, msg)
	return
}

func DecodeSession(msg ifc.Message) (rm *Session, e error) {
	return
}

type Raw struct {
	*message.Addresses
	Bytes []byte
}

func (r Raw) Encode() (o ifc.Message) {
	o, _ = message.Encode(r.To, r.From, r.Bytes)
	return
}
