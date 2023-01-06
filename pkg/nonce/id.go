package nonce

import (
	"crypto/rand"

	"github.com/indra-labs/indra/pkg/sha256"
)

const IDLen = 8

type ID [IDLen]byte

var seed sha256.Hash
var counter uint16

func reseed() {
	var c int
	var e error
	if c, e = rand.Read(seed[:]); check(e) && c != IDLen {
		panic(e)
	}
	counter++
}

func NewID() (t ID) {
	if counter == 0 {
		reseed()
	}
	s := sha256.Single(seed[:])
	copy(seed[:], s[:])
	copy(t[:], seed[:IDLen])
	return
}
