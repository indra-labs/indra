package relay

import (
	"os"
	"sync"
	"testing"
	"time"
	
	"github.com/cybriq/qu"
	"go.uber.org/atomic"
	
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
)

func TestClient_SendSessionKeys(t *testing.T) {
	log2.SetLogLevel(log2.Trace)
	var clients []*Engine
	var e error
	if clients, e = CreateNMockCircuits(2, 2); check(e) {
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
		case <-time.After(time.Second * 2):
		case <-quit:
			return
		}
		for i := 0; i < int(counter.Load()); i++ {
			wg.Done()
		}
		t.Error("SendSessionKeys test failed")
		os.Exit(1)
	}()
	for i := 0; i < 10; i++ {
		log.D.Ln("buying sessions", i)
		wg.Add(1)
		counter.Inc()
		e = clients[0].BuyNewSessions(1000000, func() {
			wg.Done()
			counter.Dec()
		})
		if check(e) {
			wg.Done()
			counter.Dec()
		}
		wg.Wait()
		quit.Q()
		for j := range clients[0].SessionCache {
			log.D.F("%d %s %v", i, j, clients[0].SessionCache[j])
		}
	}
	for _, v := range clients {
		v.Shutdown()
	}
}
