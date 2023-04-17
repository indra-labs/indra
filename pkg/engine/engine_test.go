package engine

import (
	"context"
	"runtime"
	"sync"
	"testing"
	"time"
	
	"github.com/cybriq/qu"
	"go.uber.org/atomic"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/engine/onions"
	"git-indra.lan/indra-labs/indra/pkg/engine/services"
	"git-indra.lan/indra-labs/indra/pkg/engine/sessions"
	"git-indra.lan/indra-labs/indra/pkg/engine/transport"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/util/cryptorand"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
	"git-indra.lan/indra-labs/indra/pkg/util/tests"
)

func TestClient_SendExit(t *testing.T) {
	log2.SetLogLevel(log2.Trace)
	var clients []*Engine
	var e error
	ctx, cancel := context.WithCancel(context.Background())
	if clients, e = CreateNMockCircuitsWithSessions(2, 2, ctx); fails(e) {
		t.Error(e)
		t.FailNow()
	}
	client := clients[0]
	log.D.Ln("client", client.GetLocalNodeAddressString())
	// set up forwarding port service
	const port = 3455
	sim := transport.NewByteChan(0)
	for i := range clients {
		e = clients[i].AddServiceToLocalNode(&services.Service{
			Port:      port,
			Transport: sim,
			RelayRate: 58000,
		})
		if fails(e) {
			t.Error(e)
			t.FailNow()
		}
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
		t.Error("Exit test failed")
	}()
out:
	for i := 3; i < len(clients[0].Sessions)-1; i++ {
		wg.Add(1)
		var msg slice.Bytes
		if msg, _, e = tests.GenMessage(64, "request"); fails(e) {
			t.Error(e)
			t.FailNow()
		}
		var respMsg slice.Bytes
		var respHash sha256.Hash
		if respMsg, respHash, e = tests.GenMessage(32, "response"); fails(e) {
			t.Error(e)
			t.FailNow()
		}
		bob := clients[0].Sessions[i]
		returnHops := client.Manager.GetSessionsAtHop(5)
		var alice *sessions.Data
		if len(returnHops) > 1 {
			cryptorand.Shuffle(len(returnHops), func(i, j int) {
				returnHops[i], returnHops[j] = returnHops[j],
					returnHops[i]
			})
		}
		alice = returnHops[0] // c[bob.Hop] = clients[0].Sessions[i]
		id := nonce.NewID()
		client.SendExit(port, msg, id, bob, alice, func(idd nonce.ID,
			ifc interface{}, b slice.Bytes) (e error) {
			if sha256.Single(b) != respHash {
				t.Error("failed to receive expected message")
			}
			if id != idd {
				t.Error("failed to receive expected message ID")
			}
			log.I.F("success\n\n")
			wg.Done()
			return
		})
		bb := <-clients[3].ReceiveToLocalNode(port)
		log.T.S(bb.ToBytes())
		if e = clients[3].SendFromLocalNode(port, respMsg); fails(e) {
			t.Error("fail send")
		}
		log.T.Ln("response sent")
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

func TestClient_SendGetBalance(t *testing.T) {
	log2.SetLogLevel(log2.Trace)
	var clients []*Engine
	var e error
	ctx, cancel := context.WithCancel(context.Background())
	if clients, e = CreateNMockCircuitsWithSessions(2, 2,
		ctx); fails(e) {
		t.Error(e)
		t.FailNow()
	}
	client := clients[0]
	log.D.Ln("client", client.GetLocalNodeAddressString())
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
		t.Error("SendGetBalance test failed")
	}()
out:
	for i := 1; i < len(clients[0].Sessions)-1; i++ {
		wg.Add(1)
		returnHops := client.Manager.GetSessionsAtHop(5)
		var returner *sessions.Data
		if len(returnHops) > 1 {
			cryptorand.Shuffle(len(returnHops), func(i, j int) {
				returnHops[i], returnHops[j] = returnHops[j],
					returnHops[i]
			})
		}
		returner = returnHops[0]
		clients[0].SendGetBalance(returner, clients[0].Sessions[i],
			func(cf nonce.ID, ifc interface{}, b slice.Bytes) (e error) {
				log.I.Ln("success")
				wg.Done()
				quit.Q()
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

func TestEngine_SendIntroQuery(t *testing.T) {
	log2.SetLogLevel(log2.Info)
	log2.App = ""
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
	var idPrv *crypto.Prv
	if idPrv, e = crypto.GeneratePrvKey(); fails(e) {
		return
	}
	id := nonce.NewID()
	introducerHops := client.Manager.GetSessionsAtHop(2)
	var introducer *sessions.Data
	if len(introducerHops) > 1 {
		cryptorand.Shuffle(len(introducerHops), func(i, j int) {
			introducerHops[i], introducerHops[j] = introducerHops[j], introducerHops[i]
		})
	}
	introducer = introducerHops[0]
	returnHops := client.Manager.GetSessionsAtHop(5)
	var returner *sessions.Data
	if len(returnHops) > 1 {
		cryptorand.Shuffle(len(returnHops), func(i, j int) {
			returnHops[i], returnHops[j] = returnHops[j],
				returnHops[i]
		})
	}
	returner = returnHops[0]
	svc := &services.Service{
		Port:      2345,
		RelayRate: 43523,
		Transport: transport.NewByteChan(64),
	}
	client.SendHiddenService(id, idPrv, time.Now().Add(time.Hour), returner,
		introducer, svc, func(id nonce.ID, ifc interface{},
			b slice.Bytes) (e error) {
			log.I.S("hidden service callback", id, ifc, b.ToBytes())
			return
		})
	log2.SetLogLevel(log2.Trace)
	// Now query everyone for the intro.
	idPub := crypto.DerivePub(idPrv)
	peers := clients[1:]
	log.D.Ln("client address", client.GetLocalNodeAddressString())
	for i := range peers {
		wg.Add(1)
		counter.Inc()
		log.T.Ln("peer", i)
		if len(returnHops) > 1 {
			cryptorand.Shuffle(len(returnHops), func(i, j int) {
				returnHops[i], returnHops[j] = returnHops[j], returnHops[i]
			})
		}
		if len(introducerHops) > 1 {
			cryptorand.Shuffle(len(introducerHops), func(i, j int) {
				introducerHops[i], introducerHops[j] = introducerHops[j], introducerHops[i]
			})
		}
		client.SendIntroQuery(id, idPub, introducerHops[0], returnHops[0],
			func(in *onions.Intro) {
				wg.Done()
				counter.Dec()
				log.I.Ln("success",
					in.ID, in.Key.ToBase32Abbreviated(), in.AddrPort)
			})
		wg.Wait()
	}
	quit.Q()
	cancel()
}

func TestEngine_Message(t *testing.T) {
	log2.SetLogLevel(log2.Trace)
	log2.App = ""
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
			case <-time.After(time.Second * 3):
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
	svc := &services.Service{
		Port:      2345,
		RelayRate: 43523,
		Transport: transport.NewByteChan(64),
	}
	ini := client.SendHiddenService(id, idPrv, time.Now().Add(time.Hour),
		returner, introducer, svc, func(id nonce.ID, ifc interface{},
			b slice.Bytes) (e error) {
			log.I.F("hidden service %s successfully propagated", ifc)
			wg.Done()
			counter.Dec()
			return
		})
	wg.Wait()
	time.Sleep(time.Second)
	wg.Add(1)
	counter.Inc()
	var rd *onions.Ready
	client.SendRoute(ini.Key, ini.AddrPort,
		func(id nonce.ID, ifc interface{}, b slice.Bytes) (e error) {
			rd = ifc.(*onions.Ready)
			log.D.S("route pending", rd.Address, rd.Return)
			counter.Dec()
			wg.Done()
			return
		})
	wg.Wait()
	log2.SetLogLevel(log2.Trace)
	msg, _, _ := tests.GenMessage(256, "hidden service message test")
	wg.Add(1)
	counter.Inc()
	var ms *onions.Message
	client.SendMessage(&onions.Message{
		Address: rd.Address,
		ID:      nonce.NewID(),
		Re:      rd.ID,
		Forward: rd.Return,
		Return:  MakeReplyHeader(client),
		Payload: msg,
	}, func(id nonce.ID, ifc interface{}, b slice.Bytes) (e error) {
		log.D.S("request success", id, ifc)
		ms = ifc.(*onions.Message)
		counter.Dec()
		wg.Done()
		return
	})
	wg.Wait()
	wg.Add(1)
	counter.Inc()
	client.SendMessage(&onions.Message{
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

func TestClient_SendPing(t *testing.T) {
	log2.SetLogLevel(log2.Debug)
	var clients []*Engine
	var e error
	ctx, cancel := context.WithCancel(context.Background())
	if clients, e = CreateNMockCircuitsWithSessions(1, 2,
		ctx); fails(e) {
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
		var c sessions.Circuit
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

func TestEngine_Route(t *testing.T) {
	log2.SetLogLevel(log2.Info)
	log2.App = ""
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
	svc := &services.Service{
		Port:      localPort,
		RelayRate: 43523,
		Transport: transport.NewByteChan(64),
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
	cancel()
	log.W.Ln("fin")
}

func TestClient_SendSessionKeys(t *testing.T) {
	log2.SetLogLevel(log2.Debug)
	var clients []*Engine
	var e error
	ctx, cancel := context.WithCancel(context.Background())
	if clients, e = CreateNMockCircuits(2, 2, ctx); fails(e) {
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
	cancel()
}
