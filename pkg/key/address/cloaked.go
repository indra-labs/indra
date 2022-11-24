// Package address manages encryption keys to be used with a specific
// counterparty, in a list that is used by node.Node via session.Sessions in the
// SendCache and ReceiveCache data structures.
//
// Receiver keys are the private keys that are advertised in messages to be used
// in the next reply message.
//
// Sender keys are public keys taken from received messages Receiver keys, they
// are received in a cloaked form to eliminate observer correlation and provide
// a recogniser that scans the SendCache for public keys that generate the
// matching public key in order to associate a message to a node.Node.
package address

import (
	"crypto/rand"
	"unsafe"

	"github.com/Indra-Labs/indra"
	"github.com/Indra-Labs/indra/pkg/blake3"
	"github.com/Indra-Labs/indra/pkg/key/prv"
	"github.com/Indra-Labs/indra/pkg/key/pub"
	log2 "github.com/cybriq/proc/pkg/log"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

// Cloaked is the blinded hash of a public key used to conceal a message
// public key from attackers.
type Cloaked []byte

const BlindLen = 3
const HashLen = 5
const Len = BlindLen + HashLen

func (c Cloaked) CopyBlinder() (blinder []byte) {
	blinder = make([]byte, BlindLen)
	copy(blinder, c[:BlindLen])
	return
}

// Sender is the raw bytes of a public key received in the metadata of a
// message.
type Sender struct {
	*pub.Key
}

// FromPubKey creates a Sender from a public key.
func FromPubKey(k *pub.Key) (s *Sender) {
	s = &Sender{Key: k}
	return
}

// FromBytes creates a Sender from a received public key bytes.
func FromBytes(pkb pub.Bytes) (s *Sender, e error) {
	var pk *pub.Key
	pk, e = pub.FromBytes(pkb)
	s = &Sender{Key: pk}
	return
}

// GetCloak returns a value which a receiver with the private key can
// identify the association of a message with the peer in order to retrieve the
// private key to generate the message cipher.
//
// The three byte blinding factor concatenated in front of the public key
// generates the 5 bytes at the end of the Cloaked code. In this way the
// source public key it relates to is hidden to any who don't have this public
// key, which only the parties know.
func (s Sender) GetCloak() (c Cloaked, e error) {
	blinder := make([]byte, BlindLen)
	var n int
	if n, e = rand.Read(blinder); check(e) && n != BlindLen {
		return
	}
	c = Cloak(blinder, s.Key.ToBytes())
	return
}

func Cloak(blinder []byte, key pub.Bytes) (c Cloaked) {
	b := make([]byte, BlindLen)
	copy(b, blinder)
	h := blake3.Single(append(b, key...))[:HashLen]
	c = append(b, h[:HashLen]...)
	return
}

// Receiver wraps a private key with pre-generated public key used to recognise
// and associate messages from a specific peer, the public key is sent in a
// previous message inside the encrypted payload and this structure is cached to
// identify the correct key to decrypt the message.
type Receiver struct {
	*prv.Key
	Pub *pub.Key
	pub.Bytes
}

// NewReceiver takes a private key and generates an Receiver for the address
// cache.
func NewReceiver(k *prv.Key) (a *Receiver) {
	a = &Receiver{
		Key: k,
		Pub: pub.Derive(k),
	}
	a.Bytes = a.Pub.ToBytes()
	return
}

// Match uses the cached public key and the provided blinding factor to
// match the source public key so the packet address field is only recognisable
// to the intended recipient.
func (a *Receiver) Match(r Cloaked) bool {
	if len(r) != Len {
		return false
	}
	hash := Cloak(r[:BlindLen], a.Bytes)
	return *(*string)(unsafe.Pointer(&r)) ==
		*(*string)(unsafe.Pointer(&hash))

}
