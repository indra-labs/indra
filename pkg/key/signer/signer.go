// Package signer is an implementation of an efficient method for generating new
// private keys for putting a unique ECDH cipher half on every single message
// segment and eliminating correlations in the ciphertexts of the messages.
package signer

import (
	"sync"

	"github.com/Indra-Labs/indra"
	"github.com/Indra-Labs/indra/pkg/key/prv"
	log2 "github.com/Indra-Labs/indra/pkg/log"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

type KeySet struct {
	sync.Mutex
	Base, Increment *prv.Key
}

// New creates a new KeySet which enables (relatively) fast generation of new
// private keys by using scalar addition.
func New() (first *prv.Key, ks *KeySet, e error) {
	ks = &KeySet{}
	if ks.Base, e = prv.GenerateKey(); check(e) {
		return
	}
	if ks.Increment, e = prv.GenerateKey(); check(e) {
		return
	}
	first = ks.Base
	return
}

// Next adds Increment to Base, assigns the new value to the Base and returns
// the new value.
func (ks *KeySet) Next() (n *prv.Key) {
	ks.Mutex.Lock()
	next := ks.Base.Key.Add(&ks.Increment.Key)
	ks.Base.Key = *next
	n = ks.Base
	ks.Mutex.Unlock()
	return
}
