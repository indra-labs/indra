package wire

import (
	"fmt"
	"net"
	"reflect"

	"github.com/Indra-Labs/indra/pkg/ifc"
	"github.com/Indra-Labs/indra/pkg/key/pub"
	"github.com/Indra-Labs/indra/pkg/nonce"
)

// MagicLen is 3 to make it nearly impossible that the wrong cipher will yield a
// valid Magic string as listed below.
const MagicLen = 3

type MessageMagic string

func GetMagic(s string) (mm MessageMagic) {
	return MessageMagic(s[:MagicLen])
}

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

func Serialize(message interface{}) (b ifc.Message, e error) {
	switch m := message.(type) {
	case Forward:
		b = m.Encode()
	case Exit:

	case Return:

	case Cipher:

	case Purchase:

	case Session:
		
	default:
		e = fmt.Errorf("unknown type %v", reflect.TypeOf(m))
	}
	return
}

func Deserialize(b ifc.Message) (out Message, e error) {
	mm := MessageMagic(b[:MagicLen])
	switch mm {
	case ForwardMagic:
		out, e = DecodeForward(b)
	case ExitMagic:

	case ReturnMagic:

	case CipherMagic:

	case PurchaseMagic:

	case SessionMagic:

	default:
		e = fmt.Errorf("unknown message magic '%s'", mm)
	}
	return
}

// Forward is just an IP address and a wrapper for another message.
type Forward struct {
	net.IP
	ifc.Message
}

func (rm *Forward) Encode() (o ifc.Message) {
	ipLen := len(rm.IP)
	totalLen := 1 + ipLen + len(rm.Message) + 1
	o = make(ifc.Message, totalLen)
	copy(o[:MagicLen], ForwardMagic)
	o[MagicLen] = byte(ipLen)
	copy(o[MagicLen+1:MagicLen+ipLen+2], rm.IP)
	copy(o[MagicLen+ipLen+2:], rm.Message)
	return
}

func DecodeForward(msg ifc.Message) (rm *Forward, e error) {
	ipLen := int(msg[MagicLen])
	rm = &Forward{
		IP:      net.IP(msg[MagicLen+1 : MagicLen+1+ipLen]),
		Message: msg[MagicLen+1+ipLen:],
	}
	return
}

// ToForward safely asserts the type of Message to be a Forward message.
func ToForward(msg Message) (f *Forward, e error) {
	switch t := msg.(type) {
	case *Forward:
		f = t
	default:
		e = fmt.Errorf("incorrect type returned %v expected %v",
			reflect.TypeOf(t), reflect.TypeOf(&Forward{}))
	}
	return
}

// Exit messages are the layer of a message after two Forward packets that
// provides an exit address and
type Exit struct {
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
	ifc.Message
}

func (rm *Exit) Encode() (o ifc.Message) {
	return
}

func DecodeExit(msg ifc.Message) (rm *Exit, e error) {
	return
}

// ToExit safely asserts the type of Message to be an Exit message.
func ToExit(msg Message) (x *Exit, e error) {
	switch t := msg.(type) {
	case *Exit:
		x = t
	default:
		e = fmt.Errorf("incorrect type returned %v expected %v",
			reflect.TypeOf(t), reflect.TypeOf(x))
	}
	return
}

// Return messages are distinct from Forward messages in that the header
// encryption uses a different secret than the payload. Relays identify this by
// encrypting the first 16 bytes of the header and if one of the magic bytes is
// not found
type Return struct {
	nonce.ID
	// IP is the address of the next relay in the return leg of a circuit.
	net.IP
	ifc.Message
}

func (rm *Return) Encode() (o ifc.Message) {
	return
}

func DecodeReturn(msg ifc.Message) (rm *Return, e error) {
	return
}

// ToReturn safely asserts the type of Message to be an Return message.
func ToReturn(msg Message) (rm *Return, e error) {
	switch r := msg.(type) {
	case *Return:
		rm = r
	default:
		e = fmt.Errorf("incorrect type returned %v expected %v",
			reflect.TypeOf(r), reflect.TypeOf(rm))
	}
	return
}

// Cipher delivers a public key to be used in association with a Return with the
// matching ID. This is wrapped in two Forward messages and contains two layers
// of Forward messages, ensuring that the node cannot discover where this comes
// from but allows a Return public key to be provided for a Purchase.
type Cipher struct {
	nonce.ID
	Key pub.Bytes
	Forward
}

func (rm *Cipher) Encode() (o ifc.Message) {
	return
}

func DecodeCipher(msg ifc.Message) (rm *Cipher, e error) {
	return
}

// ToCipher safely asserts the type of Message to be a Cipher message.
func ToCipher(msg Message) (c *Cipher, e error) {
	switch t := msg.(type) {
	case *Cipher:
		c = t
	default:
		e = fmt.Errorf("incorrect type returned %v expected %v",
			reflect.TypeOf(t), reflect.TypeOf(c))
	}
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
	Bytes   int
	Receipt ifc.Message
	Return
}

func (rm *Purchase) Encode() (o ifc.Message) {
	return
}

func DecodePurchase(msg ifc.Message) (rm *Purchase, e error) {
	return
}

// ToPurchase safely asserts the type of Message to be a Purchase message.
func ToPurchase(msg Message) (p *Purchase, e error) {
	switch t := msg.(type) {
	case *Purchase:
		p = t
	default:
		e = fmt.Errorf("incorrect type returned %v expected %v",
			reflect.TypeOf(t), reflect.TypeOf(p))
	}
	return
}

// Session is a message containing two public keys which identify to a relay the
// session to account bytes on, this is wrapped in two Return message layers by
// the seller. Forward keys are used for encryption in Forward and Exit
// messages, and Return keys are separate and are only known to the client and
// relay that issues a Session, ensuring that the Exit cannot see the inner
// layers of the Return messages.
type Session struct {
	Forward pub.Bytes
	Return  pub.Bytes
}

func (rm *Session) Encode() (o ifc.Message) {
	return
}

func DecodeSession(msg ifc.Message) (rm *Session, e error) {
	return
}

// ToSession safely asserts the type of Message to be a Purchase message.
func ToSession(msg Message) (s *Session, e error) {
	switch t := msg.(type) {
	case *Session:
		s = t
	default:
		e = fmt.Errorf("incorrect type returned %v expected %v",
			reflect.TypeOf(t), reflect.TypeOf(s))
	}
	return
}
