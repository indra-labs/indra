package engine

import (
	"context"
	"reflect"
	"sync"
	"testing"
	"time"

	"go.uber.org/atomic"

	"git-indra.lan/indra-labs/indra/pkg/crypto"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/engine/services"
	"git-indra.lan/indra-labs/indra/pkg/engine/sessions"
	"git-indra.lan/indra-labs/indra/pkg/engine/transport"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/util/cryptorand"
	"git-indra.lan/indra-labs/indra/pkg/util/qu"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

func TestEngine_SendHiddenService(t *testing.T) {
	log2.SetLogLevel(log2.Debug)
	log2.App.Store("")
	var clients []*Engine
	var e error
	const nCircuits = 10
	ctx, cancel := context.WithCancel(context.Background())
	if clients, e = CreateNMockCircuits(nCircuits, nCircuits,
		ctx); fails(e) {
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
			quit.Q()
			t.Error("MakeHiddenService test failed")
		case <-quit:
			for _, v := range clients {
				v.Shutdown()
			}
			for i := 0; i < int(counter.Load()); i++ {
				wg.Done()
			}
			return
		}
	}()
	for i := 0; i < nCircuits*nCircuits/2; i++ {
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
	}
	log2.SetLogLevel(log2.Debug)
	var idPrv *crypto.Prv
	if idPrv, e = crypto.GeneratePrvKey(); fails(e) {
		return
	}
	id := nonce.NewID()
	introducerHops := clients[0].Manager.GetSessionsAtHop(2)
	returnHops := clients[0].Manager.GetSessionsAtHop(5)
	var introducer *sessions.Data
	if len(introducerHops) > 1 {
		cryptorand.Shuffle(len(introducerHops), func(i, j int) {
			introducerHops[i], introducerHops[j] = introducerHops[j],
				introducerHops[i]
		})
	}
	var returner *sessions.Data
	if len(introducerHops) > 1 {
		cryptorand.Shuffle(len(returnHops), func(i, j int) {
			returnHops[i], returnHops[j] = returnHops[j],
				returnHops[i]
		})
	}
	// There must be at least one, and if there was more than one the first
	// index of introducerHops will be a randomly selected one.
	introducer = introducerHops[0]
	returner = returnHops[0]
	wg.Add(1)
	counter.Inc()
	svc := &services.Service{
		Port:      2345,
		RelayRate: 43523,
		Transport: transport.NewByteChan(64),
	}
	clients[0].SendHiddenService(id, idPrv, 0, 0,
		time.Now().Add(time.Hour), returner, introducer, svc,
		func(id nonce.ID, ifc interface{}, b slice.Bytes) (e error) {
			log.W.S("received intro", reflect.TypeOf(ifc), b.ToBytes())
			// This happens when the gossip gets back to us.
			wg.Done()
			counter.Dec()
			return
		})
	wg.Wait()
	quit.Q()
	cancel()
}
