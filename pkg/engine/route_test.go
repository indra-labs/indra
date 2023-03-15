package engine

import (
	"sync"
	"testing"
	"time"
	
	"github.com/cybriq/qu"
	"go.uber.org/atomic"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/util/cryptorand"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

func TestEngine_Route(t *testing.T) {
	log2.SetLogLevel(log2.Info)
	log2.App = "test"
	var clients []*Engine
	var e error
	const nCircuits = 5
	if clients, e = CreateNMockCircuits(nCircuits, nCircuits); check(e) {
		t.Error(e)
		t.FailNow()
	}
	client := clients[0]
	// Start up the clients.
	for _, v := range clients {
		go v.Start()
	}
	var wg sync.WaitGroup
	var counter atomic.Int32
	quit := qu.T()
	go func() {
		select {
		case <-time.After(time.Second * 20):
			quit.Q()
			t.Error("Route test failed")
		case <-quit:
			for i := 0; i < int(counter.Load()); i++ {
				wg.Done()
			}
			for _, v := range clients {
				v.Shutdown()
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
		if check(e) {
			wg.Done()
			counter.Dec()
		}
		wg.Wait()
	}
	var idPrv *prv.Key
	if idPrv, e = prv.GenerateKey(); check(e) {
		return
	}
	id := nonce.NewID()
	introducerHops := client.SessionManager.GetSessionsAtHop(2)
	var introducer *SessionData
	if len(introducerHops) > 1 {
		cryptorand.Shuffle(len(introducerHops), func(i, j int) {
			introducerHops[i], introducerHops[j] = introducerHops[j], introducerHops[i]
		})
	}
	// There must be at least one, and if there was more than one the first
	// index of introducerHops will be a randomly selected one.
	introducer = introducerHops[0]
	client.SendHiddenService(id, idPrv,
		time.Now().Add(time.Hour), introducer,
		func(id nonce.ID, k *pub.Bytes, b slice.Bytes) (e error) {
			log.I.S("hidden service callback", id, k, b.ToBytes())
			return
		})
	// Now query everyone for the intro.
	idPub := pub.Derive(idPrv)
	// peers := clients[1:]
	delete(client.Introductions.KnownIntros, idPub.ToBytes())
	returnHops := client.SessionManager.GetSessionsAtHop(2)
	for i := range returnHops {
		wg.Add(1)
		counter.Inc()
		if len(returnHops) > 1 {
			cryptorand.Shuffle(len(returnHops), func(i, j int) {
				returnHops[i], returnHops[j] = returnHops[j], returnHops[i]
			})
		}
		client.SendIntroQuery(id, idPub, returnHops[0],
			func(id nonce.ID, k *pub.Bytes, b slice.Bytes) (e error) {
				wg.Done()
				counter.Dec()
				log.I.Ln("success", i)
				return
			})
		wg.Wait()
	}
	log.I.Ln("all peers know about the hidden service")
	log2.SetLogLevel(log2.Debug)
	
	quit.Q()
}
