package engine

import (
	"net"
	"testing"
	"time"
	
	"github.com/cybriq/qu"
	
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
)

func TestRUDP(t *testing.T) {
	log2.SetLogLevel(log2.Trace)
	var e error
	var one, two *Keys
	var lis *RCP
	if one, two, e = Generate2Keys(); fails(e) {
		t.FailNow()
	}
	_ = two
	quit := qu.T()
	if lis, e = NewListenerRCP(one, net.ParseIP(""), 64, 1382, quit); fails(e) {
		t.FailNow()
	}
	time.Sleep(time.Second)
	lis.Stop()
}
