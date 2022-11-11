package address

import (
	"crypto/rand"
	"unsafe"

	"github.com/Indra-Labs/indra"
	"github.com/Indra-Labs/indra/pkg/key/prv"
	"github.com/Indra-Labs/indra/pkg/key/pub"
	"github.com/Indra-Labs/indra/pkg/sha256"
	log2 "github.com/cybriq/proc/pkg/log"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

// Recipient is the blinded hash of a public key used to conceal a message
// public key from attackers.
type Recipient []byte

// Address is the raw bytes of a public key received in the metadata of a
// message.
type Address pub.Bytes

// NewAddress creates a recipient Address from a received public key bytes.
func NewAddress(bytes pub.Bytes) (a Address) { return Address(bytes) }

// GetCloakedAddress returns a value which a receiver with the private key can
// identify the association of a message with the peer in order to retrieve the
// private key to generate the message cipher.
//
// The two byte blinding factor concatenated in front of the public key
// generates the 4 bytes at the end of the Recipient code. In this way the
// source public key it relates to is hidden to any who don't have this public
// key, which only the parties know.
func (a Address) GetCloakedAddress() (r Recipient, e error) {
	blinder := make([]byte, 2)
	var n int
	if n, e = rand.Read(blinder); check(e) && n != 2 {
		return
	}
	h := sha256.Single(append(blinder, a...))
	r = append(blinder, h[:4]...)
	return
}

// Addressee wraps a private key with pre-generated public key used to recognise
// and associate messages from a specific peer, the public key is sent in a
// previous message inside the encrypted payload and this structure is cached to
// identify the correct key to decrypt the message.
type Addressee struct {
	*prv.Key
	pub.Bytes
}

// NewAddressee takes a private key and generates an Addressee for the address
// cache.
func NewAddressee(k *prv.Key) (a *Addressee) {
	a = &Addressee{Key: k}
	pub := pub.Derive(k)
	a.Bytes = pub.ToBytes()
	return
}

// IsAddress uses the cached public key and the provided blinding factor to
// match the source public key so the packet address field is only recognisable
// to the intended recipient.
func (a *Addressee) IsAddress(r Recipient) bool {
	if len(r) != 6 {
		return false
	}
	blinder, print := r[:2], r[2:]
	hash := sha256.Single(append(blinder, a.Bytes...))[:4]
	return *(*string)(unsafe.Pointer(&print)) ==
		*(*string)(unsafe.Pointer(&hash))

}
