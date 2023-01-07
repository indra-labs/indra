package address

import (
	"math/rand"
	"testing"
	"time"

	"github.com/indra-labs/indra/pkg/key/prv"
	"github.com/indra-labs/indra/pkg/key/pub"
)

func TestAddress(t *testing.T) {
	var e error
	var sendPriv *prv.Key
	if sendPriv, e = prv.GenerateKey(); check(e) {
		return
	}
	sendPub := pub.Derive(sendPriv)
	sendBytes := sendPub.ToBytes()
	var cloaked Cloaked
	cloaked = GetCloak(sendPub)
	if !Match(cloaked, sendBytes) {
		t.Error("failed to recognise cloaked address")
	}
	rand.Seed(time.Now().Unix())
	flip := rand.Intn(Len)
	var broken Cloaked
	copy(broken[:], cloaked[:])
	broken[flip] = ^broken[flip]
	if Match(broken, sendBytes) {
		t.Error("recognised incorrectly broken cloaked address")
	}
}
