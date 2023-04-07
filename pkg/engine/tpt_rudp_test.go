package engine

import (
	"testing"
	"time"
	
	"github.com/cybriq/qu"
	
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
)

func TestRCP(t *testing.T) {
	log2.SetLogLevel(log2.Trace)
	var e error
	var one, two *Keys
	var lis *RCP
	if one, two, e = Generate2Keys(); fails(e) {
		t.FailNow()
	}
	_ = two
	quit := qu.T()
	listenerAddress := "0.0.0.0"
	if lis, e = NewListenerRCP(one, listenerAddress, 64, 1382,
		quit); fails(e) {
		t.FailNow()
	}
	
	lis.Stop()
	time.Sleep(10 * time.Second)
	<-quit
}
