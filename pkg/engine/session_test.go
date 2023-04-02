package engine

import (
	"sync"
	"testing"
	"time"
	
	"github.com/cybriq/qu"
	"go.uber.org/atomic"
	
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

func TestOnionSkins_Session(t *testing.T) {
	log2.SetLogLevel(log2.Trace)
	sess := NewSessionKeys(1)
	on := Skins{}.
		Session(sess).
		Assemble()
	s := Encode(on)
	s.SetCursor(0)
	var onc Onion
	if onc = Recognise(s, slice.GenerateRandomAddrPortIPv6()); onc == nil {
		t.Error("did not unwrap")
		t.FailNow()
	}
	if e := onc.Decode(s); fails(e) {
		t.Error("did not decode")
		t.FailNow()
		
	}
	var ci *Session
	var ok bool
	if ci, ok = onc.(*Session); !ok {
		t.Error("did not unwrap expected type")
		t.FailNow()
	}
	if !ci.Header.Key.Equals(&sess.Header.Key) {
		t.Error("header key did not unwrap correctly")
		t.FailNow()
	}
	if !ci.Payload.Key.Equals(&sess.Payload.Key) {
		t.Error("payload key did not unwrap correctly")
		t.FailNow()
	}
}

func TestClient_SendSessionKeys(t *testing.T) {
	log2.SetLogLevel(log2.Debug)
	var clients []*Engine
	var e error
	if clients, e = CreateNMockCircuits(2, 2); fails(e) {
		t.Error(e)
		t.FailNow()
	}
	// Start up the clients.
	for _, v := range clients {
		go v.Start()
	}
	var wg sync.WaitGroup
	var counter atomic.Int32
	quit := qu.T()
	go func() {
		select {
		case <-time.After(time.Second * 6):
		case <-quit:
			return
		}
		for i := 0; i < int(counter.Load()); i++ {
			wg.Done()
		}
		t.Error("SendSessionKeys test failed")
		quit.Q()
	}()
	for i := 0; i < 10; i++ {
		log.D.Ln("buying sessions", i)
		wg.Add(1)
		counter.Inc()
		e = clients[0].BuyNewSessions(1000000, func() {
			wg.Done()
			counter.Dec()
		})
		if fails(e) {
			wg.Done()
			counter.Dec()
		}
		wg.Wait()
		for j := range clients[0].SessionCache {
			log.D.F("%d %s %v", i, j, clients[0].SessionCache[j])
		}
		quit.Q()
	}
	for _, v := range clients {
		v.Shutdown()
	}
}
