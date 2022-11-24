package testutils

import (
	"crypto/rand"

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

func GenerateTestMessage(msgSize int) (msg []byte, hash blake3.Hash, e error) {
	msg = make([]byte, msgSize)
	var n int
	if n, e = rand.Read(msg); check(e) && n != msgSize {
		return
	}
	copy(msg, "payload")
	hash = blake3.Single(msg)
	return
}

func GenerateTestKeyPairs() (rp *prv.Key, rP *pub.Key, e error) {
	if rp, e = prv.GenerateKey(); check(e) {
		return
	}
	rP = pub.Derive(rp)
	return
}
