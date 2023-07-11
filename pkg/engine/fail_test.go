///go:build failingtests

package engine

import (
	"context"
	"github.com/indra-labs/indra"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/engine/sessions"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/cryptorand"
	"github.com/indra-labs/indra/pkg/util/qu"
	"github.com/indra-labs/indra/pkg/util/slice"
	"go.uber.org/atomic"
	"sync"
	"testing"
	"time"
)

//func TestEngine_Message(t *testing.T) {
//	if indra.CI == "false" {
//		log2.SetLogLevel(log2.Info)
//	}
//	var clients []*Engine
//	var e error
//	const nCircuits = 10
//	ctx, cancel := context.WithCancel(context.Background())
//	if clients, e = CreateNMockCircuits(nCircuits, nCircuits,
//		ctx); fails(e) {
//		t.Error(e)
//		t.FailNow()
//	}
//	client := clients[0]
//	log.D.Ln("client", client.Mgr().GetLocalNodeAddressString())
//	// Start up the clients.
//	for _, v := range clients {
//		go v.Start()
//	}
//	var wg sync.WaitGroup
//	var counter atomic.Int32
//	quit := qu.T()
//	go func() {
//		for {
//			select {
//			case <-time.After(time.Second * 5):
//				quit.Q()
//				t.Error("MakeHiddenService test failed")
//			case <-quit:
//				for i := 0; i < int(counter.Load()); i++ {
//					wg.Done()
//				}
//				for _, v := range clients {
//					v.Shutdown()
//				}
//				return
//			}
//		}
//	}()
//	for i := 0; i < nCircuits*nCircuits/2; i++ {
//		wg.Add(1)
//		counter.Inc()
//		e = client.BuyNewSessions(1000000, func() {
//			wg.Done()
//			counter.Dec()
//		})
//		if fails(e) {
//			wg.Done()
//			counter.Dec()
//		}
//		wg.Wait()
//	}
//	var idPrv *crypto.Prv
//	_ = idPrv
//	if idPrv, e = crypto.GeneratePrvKey(); fails(e) {
//		return
//	}
//	id := nonce.NewID()
//	_ = id
//	introducerHops := client.Mgr().GetSessionsAtHop(2)
//	var introducer *sessions.Data
//	returnHops := client.Mgr().GetSessionsAtHop(5)
//	var returner *sessions.Data
//	_ = returner
//	if len(introducerHops) > 1 {
//		cryptorand.Shuffle(len(introducerHops),
//			func(i, j int) {
//				introducerHops[i], introducerHops[j] =
//					introducerHops[j], introducerHops[i]
//			},
//		)
//	}
//	introducer = introducerHops[0]
//	if len(returnHops) > 1 {
//		cryptorand.Shuffle(len(returnHops), func(i, j int) {
//			returnHops[i], returnHops[j] = returnHops[j],
//				returnHops[i]
//		})
//	}
//	returner = returnHops[0]
//	log.D.Ln("getting sessions for introducer...")
//	for i := range clients {
//		if introducer.Node.ID == clients[i].Mgr().GetLocalNode().ID {
//			for j := 0; j < nCircuits; j++ {
//				wg.Add(1)
//				counter.Inc()
//				e = clients[i].BuyNewSessions(1000000, func() {
//					wg.Done()
//					counter.Dec()
//				})
//				if fails(e) {
//					wg.Done()
//					counter.Dec()
//				}
//			}
//			wg.Wait()
//			break
//		}
//	}
//	if indra.CI == "false" {
//		log2.SetLogLevel(log2.Trace)
//	}
//	wg.Add(1)
//	counter.Inc()
//	svc := &services.Service{
//		Port:      2345,
//		RelayRate: 43523,
//		Transport: transport.NewByteChan(64),
//	}
//	ini, _ := client.SendHiddenService(id, idPrv, 0, 0,
//		time.Now().Add(time.Hour), returner, introducer, svc,
//		func(id nonce.ID, ifc interface{}, b slice.Bytes) (e error) {
//			log.I.F("hidden service %s successfully propagated", ifc)
//			wg.Done()
//			counter.Dec()
//			return
//		})
//	wg.Wait()
//	time.Sleep(time.Second)
//	wg.Add(1)
//	counter.Inc()
//	var rd *ready.Ready
//	client.SendRoute(ini.Key, ini.AddrPort,
//		func(id nonce.ID, ifc interface{}, b slice.Bytes) (e error) {
//			rd = ifc.(*ready.Ready)
//			log.D.S("route pending", rd.Address, rd.Return)
//			counter.Dec()
//			wg.Done()
//			return
//		})
//	wg.Wait()
//	msg, _, _ := tests.GenMessage(256, "hidden service message test")
//	wg.Add(1)
//	counter.Inc()
//	var ms *whisper.Message
//	client.SendMessage(&whisper.Message{
//		Address: rd.Address,
//		ID:      nonce.NewID(),
//		Re:      rd.ID,
//		Forward: rd.Return,
//		Return:  MakeReplyHeader(client),
//		Payload: msg,
//	}, func(id nonce.ID, ifc interface{}, b slice.Bytes) (e error) {
//		log.D.S("request success", id, ifc)
//		ms = ifc.(*whisper.Message)
//		counter.Dec()
//		wg.Done()
//		return
//	})
//	wg.Wait()
//	wg.Add(1)
//	counter.Inc()
//	client.SendMessage(&whisper.Message{
//		Address: ms.Address,
//		ID:      nonce.NewID(),
//		Re:      ms.ID,
//		Forward: ms.Return,
//		Return:  MakeReplyHeader(client),
//		Payload: msg,
//	}, func(id nonce.ID, ifc interface{}, b slice.Bytes) (e error) {
//		log.D.S("response success", id, ifc)
//		counter.Dec()
//		wg.Done()
//		return
//	})
//	wg.Wait()
//	time.Sleep(time.Second)
//	quit.Q()
//	cancel()
//	log.W.Ln("fin")
//}

//func TestEngine_Route(t *testing.T) {
//	if indra.CI == "false" {
//		log2.SetLogLevel(log2.Debug)
//	}
//	runtime.GOMAXPROCS(1)
//	var clients []*Engine
//	var e error
//	const nCircuits = 10
//	ctx, cancel := context.WithCancel(context.Background())
//	if clients, e = CreateNMockCircuits(nCircuits, nCircuits,
//		ctx); fails(e) {
//		t.Error(e)
//		t.FailNow()
//	}
//	client := clients[0]
//	log.W.Ln("client", client.Mgr().GetLocalNodeAddressString())
//	// Start up the clients.
//	for _, v := range clients {
//		go v.Start()
//	}
//	var wg sync.WaitGroup
//	var counter atomic.Int32
//	quit := qu.T()
//	go func() {
//		for {
//			select {
//			case <-time.After(time.Second * 4):
//				quit.Q()
//				t.Error("Route test failed")
//			case <-quit:
//				for i := 0; i < int(counter.Load()); i++ {
//					wg.Done()
//				}
//				for _, v := range clients {
//					v.Shutdown()
//				}
//				return
//			}
//		}
//	}()
//	for i := 0; i < nCircuits*nCircuits/2; i++ {
//		wg.Add(1)
//		counter.Inc()
//		e = client.BuyNewSessions(1000000, func() {
//			wg.Done()
//			counter.Dec()
//		})
//		if fails(e) {
//			wg.Done()
//			counter.Dec()
//		}
//		wg.Wait()
//	}
//	var idPrv *crypto.Prv
//	_ = idPrv
//	if idPrv, e = crypto.GeneratePrvKey(); fails(e) {
//		return
//	}
//	id := nonce.NewID()
//	_ = id
//	introducerHops := client.Mgr().GetSessionsAtHop(2)
//	var introducer *sessions.Data
//	returnHops := client.Mgr().GetSessionsAtHop(5)
//	var returner *sessions.Data
//	_ = returner
//	if len(introducerHops) > 1 {
//		cryptorand.Shuffle(len(introducerHops),
//			func(i, j int) {
//				introducerHops[i], introducerHops[j] =
//					introducerHops[j], introducerHops[i]
//			},
//		)
//	}
//	introducer = introducerHops[0]
//	if len(returnHops) > 1 {
//		cryptorand.Shuffle(len(returnHops), func(i, j int) {
//			returnHops[i], returnHops[j] = returnHops[j],
//				returnHops[i]
//		})
//	}
//	returner = returnHops[0]
//	const localPort = 25234
//	log.D.Ln("getting sessions for introducer...")
//	for i := range clients {
//		if introducer.Node.ID == clients[i].Mgr().GetLocalNode().ID {
//			for j := 0; j < nCircuits; j++ {
//				wg.Add(1)
//				counter.Inc()
//				e = clients[i].BuyNewSessions(1000000, func() {
//					wg.Done()
//					counter.Dec()
//				})
//				if fails(e) {
//					wg.Done()
//					counter.Dec()
//				}
//			}
//			wg.Wait()
//			break
//		}
//	}
//	wg.Add(1)
//	counter.Inc()
//	svc := &services.Service{
//		Port:      localPort,
//		RelayRate: 43523,
//		Transport: transport.NewByteChan(64),
//	}
//	ini, _ := client.SendHiddenService(id, idPrv, 0, 0,
//		time.Now().Add(time.Hour),
//		returner, introducer, svc,
//		func(id nonce.ID, ifc interface{}, b slice.Bytes) (e error) {
//			log.I.F("hidden service %s successfully propagated", ifc)
//			wg.Done()
//			counter.Dec()
//			return
//		})
//	wg.Wait()
//	time.Sleep(time.Second)
//	if indra.CI == "false" {
//		log2.SetLogLevel(log2.Debug)
//	}
//	wg.Add(1)
//	counter.Inc()
//	log.D.Ln("intro", ini.ID, ini.AddrPort.String(), ini.Key.ToBased32Abbreviated(),
//		ini.Expiry, ini.Validate())
//	client.SendRoute(ini.Key, ini.AddrPort,
//		func(id nonce.ID, ifc interface{}, b slice.Bytes) (e error) {
//			log.I.S("success", id)
//			counter.Dec()
//			wg.Done()
//			return
//		})
//	wg.Wait()
//	quit.Q()
//	cancel()
//	log.W.Ln("fin")
//}

//func TestEngine_SendHiddenService(t *testing.T) {
//	if indra.CI == "false" {
//		log2.SetLogLevel(log2.Debug)
//	}
//	var clients []*Engine
//	var e error
//	const nCircuits = 10
//	ctx, cancel := context.WithCancel(context.Background())
//	if clients, e = CreateNMockCircuits(nCircuits, nCircuits,
//		ctx); fails(e) {
//		t.Error(e)
//		t.FailNow()
//	}
//	// Start up the clients.
//	for _, v := range clients {
//		go v.Start()
//	}
//	var wg sync.WaitGroup
//	var counter atomic.Int32
//	quit := qu.T()
//	go func() {
//		select {
//		case <-time.After(time.Second * 6):
//			quit.Q()
//			t.Error("MakeHiddenService test failed")
//		case <-quit:
//			for _, v := range clients {
//				v.Shutdown()
//			}
//			for i := 0; i < int(counter.Load()); i++ {
//				wg.Done()
//			}
//			return
//		}
//	}()
//	for i := 0; i < nCircuits*nCircuits/2; i++ {
//		wg.Add(1)
//		counter.Inc()
//		e = clients[0].BuyNewSessions(1000000, func() {
//			wg.Done()
//			counter.Dec()
//		})
//		if fails(e) {
//			wg.Done()
//			counter.Dec()
//		}
//		wg.Wait()
//	}
//	if indra.CI == "false" {
//		log2.SetLogLevel(log2.Debug)
//	}
//	var idPrv *crypto.Prv
//	if idPrv, e = crypto.GeneratePrvKey(); fails(e) {
//		return
//	}
//	id := nonce.NewID()
//	introducerHops := clients[0].Mgr().GetSessionsAtHop(2)
//	returnHops := clients[0].Mgr().GetSessionsAtHop(5)
//	var introducer *sessions.Data
//	if len(introducerHops) > 1 {
//		cryptorand.Shuffle(len(introducerHops), func(i, j int) {
//			introducerHops[i], introducerHops[j] = introducerHops[j],
//				introducerHops[i]
//		})
//	}
//	var returner *sessions.Data
//	if len(introducerHops) > 1 {
//		cryptorand.Shuffle(len(returnHops), func(i, j int) {
//			returnHops[i], returnHops[j] = returnHops[j],
//				returnHops[i]
//		})
//	}
//	// There must be at least one, and if there was more than one the first
//	// index of introducerHops will be a randomly selected one.
//	introducer = introducerHops[0]
//	returner = returnHops[0]
//	//wg.Add(1)
//	counter.Inc()
//	svc := &services.Service{
//		Port:      2345,
//		RelayRate: 43523,
//		Transport: transport.NewByteChan(64),
//	}
//	clients[0].SendHiddenService(id, idPrv, 0, 0,
//		time.Now().Add(time.Hour), returner, introducer, svc,
//		func(id nonce.ID, ifc interface{}, b slice.Bytes) (e error) {
//			log.W.S("received intro", reflect.TypeOf(ifc), b.ToBytes())
//			// This happens when the gossip gets back to us.
//			//wg.Done()
//			counter.Dec()
//			return
//		})
//	//wg.Wait()
//	time.Sleep(time.Second)
//	quit.Q()
//	cancel()
//}

//func TestDispatcher_Rekey(t *testing.T) {
//	if indra.CI == "false" {
//		log2.SetLogLevel(log2.Debug)
//	}
//	var e error
//	var l1, l2 *transport.Listener
//	_ = l2
//	var k1, k2 *crypto.Keys
//	ctx, cancel := context.WithCancel(context.Background())
//	_ = cancel
//	if k1, k2, e = crypto.Generate2Keys(); fails(e) {
//		t.FailNow()
//	}
//	dataPath, err := os.MkdirTemp(os.TempDir(), "badger")
//	if err != nil {
//		t.FailNow()
//	}
//	l1, e = transport.NewListener("", transport.LocalhostZeroIPv4TCP,
//		dataPath, k1, ctx, transport.DefaultMTU)
//	if fails(e) {
//		t.FailNow()
//	}
//	dataPath, err = os.MkdirTemp(os.TempDir(), "badger")
//	if err != nil {
//		t.FailNow()
//	}
//	l2, e = transport.NewListener(transport.GetHostAddress(l1.Host),
//		transport.LocalhostZeroIPv4TCP, dataPath, k2, ctx, transport.DefaultMTU)
//	if fails(e) {
//		t.FailNow()
//	}
//	var msg1, msg2 []byte
//	_ = msg2
//	msg1, _, e = tests.GenMessage(4096, "REQUEST")
//	msg2, _, e = tests.GenMessage(1024, "RESPONSE")
//	_, _ = msg1, msg2
//	hn1 := transport.GetHostAddress(l2.Host)
//	// hn2 := transport.GetHostAddress(l1.Host)
//	var ks *crypto.KeySet
//	_, ks, e = crypto.NewSigner()
//	d1 := dispatcher.NewDispatcher(l1.Dial(hn1), ctx, ks)
//	d2 := dispatcher.NewDispatcher(<-l2.Accept(), ctx, ks)
//	_, _ = d1, d2
//	var msgp1, msgp2 slice.Bytes
//	id1, id2 := nonce.NewID(), nonce.NewID()
//	on1 := ont.Assemble(Skins{
//		response.New(id1, 0, msg1, 0)})
//	on2 := ont.Assemble(Skins{
//		response.New(id2, 0, msg2, 0)})
//	s1 := ont.Encode(on1)
//	s2 := ont.Encode(on2)
//	x1 := s1.GetAll()
//	x2 := s2.GetAll()
//	xx1 := &dispatcher.Onion{x1}
//	xx2 := &dispatcher.Onion{x2}
//	sp1 := splice.New(xx1.Len())
//	sp2 := splice.New(xx2.Len())
//	if e = xx1.Encode(sp1); fails(e) {
//		t.FailNow()
//	}
//	if e = xx2.Encode(sp2); fails(e) {
//		t.FailNow()
//	}
//	countTo, succ := 1000, 0
//	go func() {
//		for {
//			select {
//			case <-ctx.Done():
//				return
//			case b := <-d1.Duplex.Receive():
//				bb, xb2 := b.ToBytes(), x2.ToBytes()
//				if string(bb) != string(xb2) {
//					t.Error("did not receive expected message")
//					return
//				} else {
//					succ++
//					continue
//				}
//			case b := <-d2.Duplex.Receive():
//				bb, xb1 := b.ToBytes(), x1.ToBytes()
//				if string(bb) != string(xb1) {
//					t.Error("did not receive expected message")
//					return
//				} else {
//					succ++
//					continue
//				}
//			}
//		}
//	}()
//	msgp1 = sp1.GetAll()
//	msgp2 = sp2.GetAll()
//	for i := 0; i < countTo; i++ {
//		d1.SendToConn(msgp1)
//		d2.SendToConn(msgp2)
//	}
//	time.Sleep(time.Second)
//	cancel()
//	if succ != countTo*3 {
//		t.Fatal("did not receive all messages correctly", succ, countTo*3)
//	}
//}

//func TestClient_SendExit(t *testing.T) {
//	if indra.CI == "false" {
//		log2.SetLogLevel(log2.Debug)
//	}
//	var clients []*Engine
//	var e error
//	ctx, cancel := context.WithCancel(context.Background())
//	if clients, e = CreateNMockCircuitsWithSessions(2, 2,
//		ctx); fails(e) {
//		t.Error(e)
//		t.FailNow()
//	}
//	client := clients[0]
//	log.D.Ln("client", client.Mgr().GetLocalNodeAddressString())
//	// set up forwarding port service
//	const port = 3455
//	sim := transport.NewByteChan(0)
//	for i := range clients {
//		e = clients[i].Mgr().AddServiceToLocalNode(&services.Service{
//			Port:      port,
//			Transport: sim,
//			RelayRate: 58000,
//		})
//		if fails(e) {
//			t.Error(e)
//			t.FailNow()
//		}
//	}
//	// Start up the clients.
//	for _, v := range clients {
//		go v.Start()
//	}
//	quit := qu.T()
//	var wg sync.WaitGroup
//	go func() {
//		select {
//		case <-time.After(time.Second):
//		case <-quit:
//			return
//		}
//		quit.Q()
//		t.Error("Exit test failed")
//	}()
//out:
//	for i := 3; i < len(clients[0].Mgr().Sessions)-1; i++ {
//		wg.Add(1)
//		var msg slice.Bytes
//		if msg, _, e = tests.GenMessage(64, "request"); fails(e) {
//			t.Error(e)
//			t.FailNow()
//		}
//		var respMsg slice.Bytes
//		var respHash sha256.Hash
//		if respMsg, respHash, e = tests.GenMessage(32,
//			"response"); fails(e) {
//			t.Error(e)
//			t.FailNow()
//		}
//		bob := clients[0].Mgr().Sessions[i]
//		returnHops := client.Mgr().GetSessionsAtHop(5)
//		var alice *sessions.Data
//		if len(returnHops) > 1 {
//			cryptorand.Shuffle(len(returnHops), func(i, j int) {
//				returnHops[i], returnHops[j] = returnHops[j],
//					returnHops[i]
//			})
//		}
//		alice = returnHops[0] // c[bob.Hop] = clients[0].Sessions[i]
//		id := nonce.NewID()
//		client.SendExit(port, msg, id, bob, alice, func(idd nonce.ID,
//			ifc interface{}, b slice.Bytes) (e error) {
//			if sha256.Single(b) != respHash {
//				t.Error("failed to receive expected message")
//			}
//			if id != idd {
//				t.Error("failed to receive expected message Keys")
//			}
//			log.D.F("success\n\n")
//			wg.Done()
//			return
//		})
//		bb := <-clients[3].Mgr().GetLocalNode().ReceiveFrom(port)
//		log.T.S(bb.ToBytes())
//		if e = clients[3].Mgr().SendFromLocalNode(port, respMsg); fails(e) {
//			t.Error("fail send")
//		}
//		log.T.Ln("response sent")
//		select {
//		case <-quit:
//			break out
//		default:
//		}
//		wg.Wait()
//	}
//	quit.Q()
//	cancel()
//	for _, v := range clients {
//		v.Shutdown()
//	}
//}

//func TestClient_SendPing(t *testing.T) {
//	if indra.CI == "false" {
//		log2.SetLogLevel(log2.Debug)
//	}
//	var clients []*Engine
//	var e error
//	ctx, cancel := context.WithCancel(context.Background())
//	if clients, e = CreateNMockCircuitsWithSessions(1, 2,
//		ctx); fails(e) {
//		t.Error(e)
//		t.FailNow()
//	}
//	// Start up the clients.
//	for _, v := range clients {
//		go v.Start()
//	}
//	quit := qu.T()
//	var wg sync.WaitGroup
//	go func() {
//		select {
//		case <-time.After(time.Second):
//		case <-quit:
//			return
//		}
//		quit.Q()
//		t.Error("SendPing test failed")
//	}()
//out:
//	for i := 3; i < len(clients[0].Mgr().Sessions)-1; i++ {
//		wg.Add(1)
//		var c sessions.Circuit
//		sess := clients[0].Mgr().Sessions[i]
//		c[sess.Hop] = clients[0].Mgr().Sessions[i]
//		clients[0].SendPing(c,
//			func(id nonce.ID, ifc interface{}, b slice.Bytes) (e error) {
//				log.D.Ln("success")
//				wg.Done()
//				return
//			})
//		select {
//		case <-quit:
//			break out
//		default:
//		}
//		wg.Wait()
//	}
//	quit.Q()
//	cancel()
//	for _, v := range clients {
//		v.Shutdown()
//	}
//}

func TestClient_SendSessionKeys(t *testing.T) {
	if indra.CI == "false" {
		log2.SetLogLevel(log2.Debug)
	}
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
		for j := range clients[0].Mgr().CircuitCache {
			log.D.F("%d %s %v", i, j, clients[0].Mgr().CircuitCache[j])
		}
		quit.Q()
	}
	for _, v := range clients {
		v.Shutdown()
	}
	cancel()
}

func TestClient_SendGetBalance(t *testing.T) {
	if indra.CI == "false" {
		log2.SetLogLevel(log2.Trace)
	}
	var clients []*Engine
	var e error
	ctx, cancel := context.WithCancel(context.Background())
	if clients, e = CreateNMockCircuitsWithSessions(4, 4,
		ctx); fails(e) {
		t.Error(e)
		t.FailNow()
	}
	client := clients[0]
	log.D.Ln("client", client.Mgr().GetLocalNodeAddressString())
	// Start up the clients.
	for _, v := range clients {
		go v.Start()
	}
	quit := qu.T()
	var wg sync.WaitGroup
	go func() {
		select {
		case <-time.After(5 * time.Second):
		case <-quit:
			return
		}
		cancel()
		quit.Q()
		t.Error("SendGetBalance test failed")
	}()
	i := 0
	wg.Add(1)
	returnHops := client.Mgr().GetSessionsAtHop(5)
	var returner *sessions.Data
	if len(returnHops) > 1 {
		cryptorand.Shuffle(len(returnHops), func(i, j int) {
			returnHops[i], returnHops[j] = returnHops[j],
				returnHops[i]
		})
	}
	returner = returnHops[0]
	clients[0].SendGetBalance(returner, clients[0].Mgr().Sessions[i],
		func(cf nonce.ID, ifc interface{}, b slice.Bytes) (e error) {
			log.D.Ln("success")
			wg.Done()
			quit.Q()
			return
		})
	wg.Wait()
	quit.Q()
	cancel()
	for _, v := range clients {
		v.Shutdown()
	}
}
