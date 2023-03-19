package ngin

import (
	"sync"
	"testing"
	"time"
	
	"github.com/cybriq/qu"
	"github.com/gookit/color"
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
	iH := client.SessionManager.GetSessionsAtHop(2)
	var introducer *SessionData
	if len(iH) > 1 {
		cryptorand.Shuffle(len(iH),
			func(i, j int) { iH[i], iH[j] = iH[j], iH[i] },
		)
	}
	// There must be at least one, and if there was more than one the first
	// index of iH will be a randomly selected one.
	const localPort = 25234
	introducer = iH[0]
	var introClient *Engine
	log.I.F("introducer %s", color.Yellow.Sprint(introducer.AddrPort.String()))
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
	client.SendHiddenService(make(slice.Bytes, RoutingHeaderLen), id, idPrv,
		time.Now().Add(time.Hour), introducer, localPort,
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
		if len(rH) > 1 {
			cryptorand.Shuffle(len(rH), func(i, j int) {
				rH[i], rH[j] = rH[j], rH[i]
			})
		}
		client.SendIntroQuery(id, idPub, rH[0], func(in *Intro) {
			wgdec()
			ini = in
			if ini == nil {
				t.Error("got empty intro query answer")
				t.FailNow()
			}
		})
	}
	wg.Wait()
	time.Sleep(time.Second * 2)
	log.I.Ln("all peers know about the hidden service")
	log.I.S("introclient", introClient.HiddenRouting.HiddenServices,
		introClient.HiddenRouting.MyIntros, introClient.HiddenRouting.KnownIntros)
	log2.SetLogLevel(log2.Trace)
	log.D.Ln("intro", ini.ID, ini.AddrPort.String(), ini.Key.ToBase32Abbreviated(),
		ini.Expiry, ini.Validate())
	client.SendRoute(ini.Key, ini.AddrPort,
		func(id nonce.ID, k *pub.Bytes, b slice.Bytes) (e error) {
			log.I.S("success", id, k, b.ToBytes())
			return
		})
	time.Sleep(time.Second * 2)
	quit.Q()
}
