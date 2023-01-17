package tests

import (
	"crypto/rand"

	"github.com/indra-labs/indra"
	"github.com/indra-labs/indra/pkg/crypto/key/prv"
	"github.com/indra-labs/indra/pkg/crypto/key/pub"
	"github.com/indra-labs/indra/pkg/crypto/sha256"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

func GenMessage(msgSize int, hrp string) (msg []byte, hash sha256.Hash, e error) {
	msg = make([]byte, msgSize)
	var n int
	if n, e = rand.Read(msg); check(e) && n != msgSize {
		return
	}
	if hrp == "" {
		hrp = "payload"
	}
	copy(msg, hrp)
	hash = sha256.Single(msg)
	return
}

func GenerateTestKeyPairs() (sp, rp *prv.Key, sP, rP *pub.Key, e error) {
	sp, sP, e = GenerateTestKeyPair()
	rp, rP, e = GenerateTestKeyPair()
	return
}

func GenerateTestKeyPair() (sp *prv.Key, sP *pub.Key, e error) {
	if sp, e = prv.GenerateKey(); check(e) {
		return
	}
	sP = pub.Derive(sp)
	return
}
