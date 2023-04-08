package engine

import (
	"fmt"
	"testing"
	"time"
	
	"github.com/cybriq/qu"
	"github.com/davecgh/go-spew/spew"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/signer"
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
	var ks *signer.KeySet
	_, ks, e = signer.New()
	listenerAddress := "127.0.0.1"
	if lis, e = NewListenerRCP(one, listenerAddress, 64, 1382, ks,
		quit); fails(e) {
		t.FailNow()
	}
	var outbound *RCP
	go func() {
		log.D.Ln("listening")
		b := <-lis.Receive()
		log.I.S("bytes", b)
		fmt.Println("bytes", spew.Sdump(b))
		time.Sleep(time.Second)
	}()
	time.Sleep(time.Second)
	outbound, e = NewOutboundRCP(two, lis.uConn.LocalAddr().String(), one.Pub,
		"127.0.0.1:0", 64, 1382, ks, quit)
	// var msg []byte
	// if msg, _, e = tests.GenMessage(256, "b0r0k"); fails(e) {
	// 	t.FailNow()
	// }
	// time.Sleep(time.Second)
	// if e = outbound.Send(msg); fails(e) {
	// 	t.FailNow()
	// }
	time.Sleep(time.Second)
	lis.Stop()
	outbound.Stop()
	<-quit
}
