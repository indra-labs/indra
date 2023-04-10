package tests

import (
	"crypto/rand"
	
	"git-indra.lan/indra-labs/indra"
	"git-indra.lan/indra-labs/indra/pkg/crypto"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
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

func GenerateTestKeyPairs() (sp, rp *crypto.Prv, sP, rP *crypto.Pub, e error) {
	sp, sP, e = GenerateTestKeyPair()
	rp, rP, e = GenerateTestKeyPair()
	return
}

func GenerateTestKeyPair() (sp *crypto.Prv, sP *crypto.Pub, e error) {
	if sp, e = crypto.GeneratePrvKey(); check(e) {
		return
	}
	sP = crypto.DerivePub(sp)
	return
}
