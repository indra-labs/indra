//go:build failingtests

package engine

import (
	"context"
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/engine/services"
	"github.com/indra-labs/indra/pkg/engine/sessions"
	"github.com/indra-labs/indra/pkg/engine/transport"
	"github.com/indra-labs/indra/pkg/onions/message"
	"github.com/indra-labs/indra/pkg/onions/ready"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/cryptorand"
	"github.com/indra-labs/indra/pkg/util/qu"
	"github.com/indra-labs/indra/pkg/util/slice"
	"github.com/indra-labs/indra/pkg/util/tests"
	"go.uber.org/atomic"
	"reflect"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestEngine_Message(t *testing.T) {
	log2.SetLogLevel(log2.Info)
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
	client := clients[0]
	log.W.Ln("client", client.Manager.GetLocalNodeAddressString())
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
			case <-time.After(time.Second * 5):
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
	introducerHops := client.Manager.GetSessionsAtHop(2)
	var introducer *sessions.Data
	returnHops := client.Manager.GetSessionsAtHop(5)
	var returner *sessions.Data
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
	log.D.Ln("getting sessions for introducer...")
	for i := range clients {
		if introducer.Node.ID == clients[i].Manager.GetLocalNode().ID {
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
	log2.SetLogLevel(log2.Trace)
	wg.Add(1)
	counter.Inc()
	svc := &services.Service{
		Port:      2345,
		RelayRate: 43523,
		Transport: transport.NewByteChan(64),
	}
	ini := client.SendHiddenService(id, idPrv, 0, 0,
		time.Now().Add(time.Hour), returner, introducer, svc,
		func(id nonce.ID, ifc interface{}, b slice.Bytes) (e error) {
			log.I.F("hidden service %s successfully propagated", ifc)
			wg.Done()
			counter.Dec()
			return
		})
	wg.Wait()
	time.Sleep(time.Second)
	wg.Add(1)
	counter.Inc()
	var rd *ready.Ready
	client.SendRoute(ini.Key, ini.AddrPort,
		func(id nonce.ID, ifc interface{}, b slice.Bytes) (e error) {
			rd = ifc.(*ready.Ready)
			log.D.S("route pending", rd.Address, rd.Return)
			counter.Dec()
			wg.Done()
			return
		})
	wg.Wait()
	msg, _, _ := tests.GenMessage(256, "hidden service message test")
	wg.Add(1)
	counter.Inc()
	var ms *message.Message
	client.SendMessage(&message.Message{
		Address: rd.Address,
		ID:      nonce.NewID(),
		Re:      rd.ID,
		Forward: rd.Return,
		Return:  MakeReplyHeader(client),
		Payload: msg,
	}, func(id nonce.ID, ifc interface{}, b slice.Bytes) (e error) {
		log.D.S("request success", id, ifc)
		ms = ifc.(*message.Message)
		counter.Dec()
		wg.Done()
		return
	})
	wg.Wait()
	wg.Add(1)
	counter.Inc()
	client.SendMessage(&message.Message{
		Address: ms.Address,
		ID:      nonce.NewID(),
		Re:      ms.ID,
		Forward: ms.Return,
		Return:  MakeReplyHeader(client),
		Payload: msg,
	}, func(id nonce.ID, ifc interface{}, b slice.Bytes) (e error) {
		log.D.S("response success", id, ifc)
		counter.Dec()
		wg.Done()
		return
	})
	wg.Wait()
	time.Sleep(time.Second)
	quit.Q()
	cancel()
	log.W.Ln("fin")
}

func TestEngine_Route(t *testing.T) {
	log2.SetLogLevel(log2.Debug)
	log2.App.Store("")
	runtime.GOMAXPROCS(1)
	var clients []*Engine
	var e error
	const nCircuits = 10
	ctx, cancel := context.WithCancel(context.Background())
	if clients, e = CreateNMockCircuits(nCircuits, nCircuits,
		ctx); fails(e) {
		t.Error(e)
		t.FailNow()
	}
	client := clients[0]
	log.W.Ln("client", client.Manager.GetLocalNodeAddressString())
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
	introducerHops := client.Manager.GetSessionsAtHop(2)
	var introducer *sessions.Data
	returnHops := client.Manager.GetSessionsAtHop(5)
	var returner *sessions.Data
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
		if introducer.Node.ID == clients[i].Manager.GetLocalNode().ID {
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
	svc := &services.Service{
		Port:      localPort,
		RelayRate: 43523,
		Transport: transport.NewByteChan(64),
	}
	ini := client.SendHiddenService(id, idPrv, 0, 0,
		time.Now().Add(time.Hour),
		returner, introducer, svc,
		func(id nonce.ID, ifc interface{}, b slice.Bytes) (e error) {
			log.I.F("hidden service %s successfully propagated", ifc)
			wg.Done()
			counter.Dec()
			return
		})
	wg.Wait()
	time.Sleep(time.Second)
	log2.SetLogLevel(log2.Debug)
	wg.Add(1)
	counter.Inc()
	log.D.Ln("intro", ini.ID, ini.AddrPort.String(), ini.Key.ToBased32Abbreviated(),
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
	cancel()
	log.W.Ln("fin")
}

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
	//wg.Add(1)
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
			//wg.Done()
			counter.Dec()
			return
		})
	//wg.Wait()
	time.Sleep(time.Second)
	quit.Q()
	cancel()
}
