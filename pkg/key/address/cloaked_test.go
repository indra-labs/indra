package address

import (
	"math/rand"
	"testing"
	"time"

	"github.com/Indra-Labs/indra/pkg/key/prv"
)

func TestAddress(t *testing.T) {
	var e error
	var sendPriv *prv.Key
	if sendPriv, e = prv.GenerateKey(); check(e) {
		return
	}
	r := NewReceiver(sendPriv)
	s := FromPub(r.Pub)
	var cloaked Cloaked
	cloaked = s.GetCloak()
	if !r.Match(cloaked) {
		t.Error("failed to recognise cloaked address")
	}
	rand.Seed(time.Now().Unix())
	flip := rand.Intn(Len)
	var broken Cloaked
	copy(broken[:], cloaked[:])
	broken[flip] = ^broken[flip]
	if r.Match(broken) {
		t.Error("recognised incorrectly broken cloaked address")
	}
}
