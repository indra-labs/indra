package engine

import (
	"context"
	"sync"
	"testing"
	"time"
	
	"github.com/cybriq/qu"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

func TestClient_SendPing(t *testing.T) {
	log2.SetLogLevel(log2.Debug)
	var clients []*Engine
	var e error
	ctx, cancel := context.WithCancel(context.Background())
	if clients, e = CreateNMockCircuitsWithSessions(1, 2, ctx); fails(e) {
		t.Error(e)
		t.FailNow()
	}
	// Start up the clients.
	for _, v := range clients {
		go v.Start()
	}
	quit := qu.T()
	var wg sync.WaitGroup
	go func() {
		select {
		case <-time.After(time.Second):
		case <-quit:
			return
		}
		quit.Q()
		t.Error("SendPing test failed")
	}()
out:
	for i := 3; i < len(clients[0].Sessions)-1; i++ {
		wg.Add(1)
		var c Circuit
		sess := clients[0].Sessions[i]
		c[sess.Hop] = clients[0].Sessions[i]
		clients[0].SendPing(c,
			func(id nonce.ID, ifc interface{}, b slice.Bytes) (e error) {
				log.I.Ln("success")
				wg.Done()
				return
			})
		select {
		case <-quit:
			break out
		default:
		}
		wg.Wait()
	}
	quit.Q()
	cancel()
	for _, v := range clients {
		v.Shutdown()
	}
}
