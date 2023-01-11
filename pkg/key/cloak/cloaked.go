// Package cloak provides a cover for the public keys for which a node has a
// private key that prevents matching the keys between one message and another.
package cloak

import (
	"crypto/rand"

	"github.com/indra-labs/indra"
	"github.com/indra-labs/indra/pkg/key/pub"
	log2 "github.com/indra-labs/indra/pkg/log"
	"github.com/indra-labs/indra/pkg/sha256"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

const BlindLen = 3
const HashLen = 5
const Len = BlindLen + HashLen

func (c PubKey) CopyBlinder() (blinder Blinder) {
	copy(blinder[:], c[:BlindLen])
	return
}

// PubKey is the blinded hash of a public key used to conceal a message public
// key from attackers.
type PubKey [Len]byte

type Blinder [BlindLen]byte
type Hash [HashLen]byte

// GetCloak returns a value which a receiver with the private key can identify
// the association of a message with the peer in order to retrieve the private
// key to generate the message cipher.
//
// The three byte blinding factor concatenated in front of the public key
// generates the 5 bytes at the end of the PubKey code. In this way the source
// public key it relates to is hidden to any who don't have this public key,
// which only the parties know.
func GetCloak(s *pub.Key) (c PubKey) {
	var blinder Blinder
	var n int
	var e error
	if n, e = rand.Read(blinder[:]); check(e) && n != BlindLen {
		panic("no entropy")
	}
	c = Cloak(blinder, s.ToBytes())
	return
}

func Cloak(b Blinder, key pub.Bytes) (c PubKey) {
	h := sha256.Single(append(b[:], key[:]...))
	copy(c[:BlindLen], b[:BlindLen])
	copy(c[BlindLen:BlindLen+HashLen], h[:HashLen])
	return
}

// Match uses the cached public key and the provided blinding factor to match
// the source public key so the packet address field is only recognisable to the
// intended recipient.
func Match(r PubKey, k pub.Bytes) bool {
	var b Blinder
	copy(b[:], r[:BlindLen])
	hash := Cloak(b, k)
	return r == hash
}
