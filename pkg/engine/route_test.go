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
	const nCircuits = 10
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
		case <-time.After(time.Second * 4):
			quit.Q()
			t.Error("MakeHiddenService test failed")
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
	log2.SetLogLevel(log2.Info)
	var idPrv *prv.Key
	if idPrv, e = prv.GenerateKey(); check(e) {
		return
	}
	id := nonce.NewID()
	introHops := client.SessionManager.GetSessionsAtHop(2)
	var introducer *SessionData
	if len(introHops) > 1 {
		cryptorand.Shuffle(len(introHops), func(i, j int) {
			introHops[i], introHops[j] = introHops[j], introHops[i]
		})
	}
	// There must be at least one, and if there was more than one the first
	// index of introHops will be a randomly selected one.
	introducer = introHops[0]
	client.SendHiddenService(id, idPrv,
		time.Now().Add(time.Hour), introducer,
		func(id nonce.ID, b slice.Bytes) {
			log.D.Ln("yay")
		})
	for i := range clients {
		log.D.S("known intros", clients[i].KnownIntros)
	}
	log2.SetLogLevel(log2.Trace)
	// Now query everyone for the intro.
	idPub := pub.Derive(idPrv)
	log.D.Ln("client address", client.GetLocalNodeAddress())
	delete(client.Introductions.KnownIntros, idPub.ToBytes())
	wg.Add(1)
	counter.Inc()
	returnHops := client.SessionManager.GetSessionsAtHop(2)
	if len(returnHops) > 1 {
		cryptorand.Shuffle(len(returnHops), func(i, j int) {
			returnHops[i], returnHops[j] = returnHops[j], returnHops[i]
		})
	}
	client.SendIntroQuery(id, idPub, returnHops[0],
		func(id nonce.ID, b slice.Bytes) {
			wg.Done()
			counter.Dec()
			log.I.Ln("success")
			quit.Q()
		})
	wg.Wait()
	quit.Q()
	log.D.Ln("-------------------------------------")
}
