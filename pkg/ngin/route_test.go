package ngin

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
	log2.App = ""
	var clients []*Engine
	var e error
	const nCircuits = 10
	if clients, e = CreateNMockCircuits(nCircuits, nCircuits); check(e) {
		t.Error(e)
		t.FailNow()
	}
	client := clients[0]
	log.I.Ln("client at", client.GetLocalNodeAddressString())
	// Start up the clients.
	for _, v := range clients {
		go v.Start()
	}
	var wg sync.WaitGroup
	var counter atomic.Int32
	wgdec := func() {
		wg.Done()
		counter.Dec()
	}
	wginc := func() {
		wg.Add(1)
		counter.Inc()
	}
	quit := qu.T()
	go func() {
		for {
			select {
			case <-time.After(time.Second * 5):
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
		}
	}()
	for i := 0; i < nCircuits; i++ {
		wginc()
		e = clients[0].BuyNewSessions(1000000, func() {
			wgdec()
		})
		if check(e) {
			wgdec()
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
	returnHops := client.SessionManager.GetSessionsAtHop(5)
	var returner *SessionData
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
	var introClient *Engine
	// log.I.F("introducer %s", color.Yellow.Sprint(introducer.AddrPort.String()))
	log.D.Ln("getting sessions for introducer...")
	for i := range clients {
		if introducer.Node.ID == clients[i].GetLocalNode().ID {
			introClient = clients[i]
			for j := 0; j < nCircuits; j++ {
				wginc()
				e = clients[i].BuyNewSessions(1000000, func() {
					wgdec()
				})
				if check(e) {
					wgdec()
				}
			}
			wg.Wait()
			break
		}
	}
	wginc()
	client.SendHiddenService(id, idPrv, time.Now().Add(time.Hour), returner,
		introducer, localPort,
		func(id nonce.ID, k *pub.Bytes, b slice.Bytes) (e error) {
			log.I.S("hidden service callback", client.GetLocalNodeAddressString(),
				id, k, b.ToBytes())
			wgdec()
			return
		})
	wg.Wait()
	// Now query everyone for the intro.
	idPub := pub.Derive(idPrv)
	delete(client.HiddenRouting.KnownIntros, idPub.ToBytes())
	rH := client.SessionManager.GetSessionsAtHop(2)
	var ini *Intro
	for _ = range rH {
		wg.Add(1)
		counter.Inc()
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
		client.SendIntroQuery(id, idPub, introducer, returner, func(in *Intro) {
			wgdec()
			ini = in
			if ini == nil {
				t.Error("got empty intro query answer")
				t.FailNow()
			}
		})
	}
	wg.Wait()
	wg.Add(1)
	log.I.Ln("all peers know about the hidden service")
	log.I.S("introclient", introClient.HiddenRouting.HiddenServices,
		introClient.HiddenRouting.MyIntros, introClient.HiddenRouting.KnownIntros)
	log2.SetLogLevel(log2.Trace)
	log.D.Ln("intro", ini.ID, ini.AddrPort.String(), ini.Key.ToBase32Abbreviated(),
		ini.Expiry, ini.Validate())
	client.SendRoute(ini.Key, ini.AddrPort,
		func(id nonce.ID, k *pub.Bytes, b slice.Bytes) (e error) {
			log.I.S("success", id, k, b.ToBytes())
			wg.Done()
			return
		})
	wg.Wait()
	quit.Q()
}
