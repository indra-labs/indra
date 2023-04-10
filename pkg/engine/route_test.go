package engine

import (
	"runtime"
	"sync"
	"testing"
	"time"
	
	"github.com/cybriq/qu"
	"go.uber.org/atomic"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/util/cryptorand"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

func TestEngine_Route(t *testing.T) {
	log2.SetLogLevel(log2.Info)
	log2.App = ""
	runtime.GOMAXPROCS(1)
	var clients []*Engine
	var e error
	const nCircuits = 10
	if clients, e = CreateNMockCircuits(nCircuits, nCircuits); fails(e) {
		t.Error(e)
		t.FailNow()
	}
	client := clients[0]
	log.W.Ln("client", client.GetLocalNodeAddressString())
	// Start up the clients.
	for _, v := range clients {
		go v.Start()
	}
	var wg sync.WaitGroup
	var counter atomic.Int32
	quit := qu.T()
	go func() {
		for {
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
		}
	}()
	for i := 0; i < nCircuits*nCircuits/2; i++ {
		wg.Add(1)
		counter.Inc()
		e = client.BuyNewSessions(1000000, func() {
			wg.Done()
			counter.Dec()
		})
		if fails(e) {
			wg.Done()
			counter.Dec()
		}
		wg.Wait()
	}
	var idPrv *crypto.Prv
	_ = idPrv
	if idPrv, e = crypto.GeneratePrvKey(); fails(e) {
		return
	}
	id := nonce.NewID()
	_ = id
	introducerHops := client.SessionManager.GetSessionsAtHop(2)
	var introducer *SessionData
	returnHops := client.SessionManager.GetSessionsAtHop(5)
	var returner *SessionData
	_ = returner
	if len(introducerHops) > 1 {
		cryptorand.Shuffle(len(introducerHops),
			func(i, j int) {
				introducerHops[i], introducerHops[j] =
					introducerHops[j], introducerHops[i]
			},
		)
	}
	introducer = introducerHops[0]
	if len(returnHops) > 1 {
		cryptorand.Shuffle(len(returnHops), func(i, j int) {
			returnHops[i], returnHops[j] = returnHops[j],
				returnHops[i]
		})
	}
	returner = returnHops[0]
	const localPort = 25234
	log.D.Ln("getting sessions for introducer...")
	for i := range clients {
		if introducer.Node.ID == clients[i].GetLocalNode().ID {
			for j := 0; j < nCircuits; j++ {
				wg.Add(1)
				counter.Inc()
				e = clients[i].BuyNewSessions(1000000, func() {
					wg.Done()
					counter.Dec()
				})
				if fails(e) {
					wg.Done()
					counter.Dec()
				}
			}
			wg.Wait()
			break
		}
	}
	wg.Add(1)
	counter.Inc()
	svc := &Service{
		Port:      localPort,
		RelayRate: 43523,
		Transport: NewSim(64),
	}
	ini := client.SendHiddenService(id, idPrv, time.Now().Add(time.Hour),
		returner, introducer, svc, func(id nonce.ID, ifc interface{},
			b slice.Bytes) (e error) {
			log.I.F("hidden service %s successfully propagated", ifc)
			wg.Done()
			counter.Dec()
			return
		})
	_ = ini
	wg.Wait()
	time.Sleep(time.Second)
	log2.SetLogLevel(log2.Trace)
	wg.Add(1)
	counter.Inc()
	log.D.Ln("intro", ini.ID, ini.AddrPort.String(), ini.Key.ToBase32Abbreviated(),
		ini.Expiry, ini.Validate())
	client.SendRoute(ini.Key, ini.AddrPort,
		func(id nonce.ID, ifc interface{}, b slice.Bytes) (e error) {
			log.I.S("success", id)
			counter.Dec()
			wg.Done()
			return
		})
	wg.Wait()
	quit.Q()
	log.W.Ln("fin")
}
