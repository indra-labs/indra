package testutils

import (
	"crypto/rand"

	"github.com/Indra-Labs/indra"
	"github.com/Indra-Labs/indra/pkg/key/prv"
	"github.com/Indra-Labs/indra/pkg/key/pub"
	log2 "github.com/Indra-Labs/indra/pkg/log"
	"github.com/Indra-Labs/indra/pkg/sha256"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

func GenerateTestMessage(msgSize int) (msg []byte, hash sha256.Hash, e error) {
	msg = make([]byte, msgSize)
	var n int
	if n, e = rand.Read(msg); check(e) && n != msgSize {
		return
	}
	copy(msg, "payload")
	hash = sha256.Single(msg)
	return
}

func GenerateTestKeyPairs() (sp, rp *prv.Key, sP, rP *pub.Key, e error) {
	if sp, e = prv.GenerateKey(); check(e) {
		return
	}
	sP = pub.Derive(sp)
	if rp, e = prv.GenerateKey(); check(e) {
		return
	}
	rP = pub.Derive(rp)
	return
}
