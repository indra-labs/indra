// Package tests provides some helpers for tests.
package tests

import (
	"crypto/rand"

	"github.com/indra-labs/indra/pkg/crypto/sha256"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
)

var (
	log   = log2.GetLogger()
	fails = log.E.Chk
)

func GenMessage(msgSize int, hrp string) (msg []byte, hash sha256.Hash, e error) {
	msg = make([]byte, msgSize)
	var n int
	if n, e = rand.Read(msg); fails(e) && n != msgSize {
		return
	}
	if hrp == "" {
		hrp = "payload"
	}
	copy(msg, hrp)
	hash = sha256.Single(msg)
	return
}
