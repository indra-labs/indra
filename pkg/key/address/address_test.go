package address

import (
	"math/rand"
	"testing"

	"github.com/Indra-Labs/indra/pkg/key/prv"
)

func TestAddress(t *testing.T) {
	var e error
	var sendPriv *prv.Key
	if sendPriv, e = prv.GenerateKey(); check(e) {
		return
	}
	ae := NewAddressee(sendPriv)
	apk := ae.Bytes
	a := NewAddress(apk)
	var cloaked Recipient
	cloaked, e = a.GetCloakedAddress()
	if !ae.IsAddress(cloaked) {
		t.Error("failed to recognise cloaked address")
	}
	flip := rand.Intn(6)
	cloaked[flip] = ^cloaked[flip]
	if ae.IsAddress(cloaked) {
		t.Error("recognised incorrectly broken cloaked address")
	}

}
