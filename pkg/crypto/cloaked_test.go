package crypto

import (
	"math/rand"
	"testing"
	"time"
)

func TestCloakedPubKey(t *testing.T) {
	var e error
	var sendPriv *Prv
	if sendPriv, e = GeneratePrvKey(); fails(e) {
		return
	}
	sendPub := DerivePub(sendPriv)
	sendBytes := sendPub.ToBytes()
	var cloaked CloakedPubKey
	cloaked = GetCloak(sendPub)
	if !Match(cloaked, sendBytes) {
		t.Error("failed to recognise cloaked address")
	}
	rand.Seed(time.Now().Unix())
	flip := rand.Intn(CloakLen)
	var broken CloakedPubKey
	copy(broken[:], cloaked[:])
	broken[flip] = ^broken[flip]
	if Match(broken, sendBytes) {
		t.Error("recognised incorrectly broken cloaked address")
	}
}
