package relay

import (
	"os"
	"sync"
	"testing"
	"time"
	
	"github.com/cybriq/qu"
	"go.uber.org/atomic"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/service"
	"git-indra.lan/indra-labs/indra/pkg/traffic"
	"git-indra.lan/indra-labs/indra/pkg/transport"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
	"git-indra.lan/indra-labs/indra/pkg/util/tests"
)

// todo: add accounting validation to these tests where relevant
//  (check relay and client see the same balance after the operations)

func TestClient_SendSessionKeys(t *testing.T) {
	log2.SetLogLevel(log2.Debug)
	var clients []*Engine
	var e error
	if clients, e = CreateNMockCircuits(false, 2); check(e) {
		t.Error(e)
		t.FailNow()
	}
	// Start up the clients.
	for _, v := range clients {
		go v.Start()
	}
	var wg sync.WaitGroup
	var counter atomic.Int32
	go func() {
		select {
		case <-time.After(time.Second * 2):
		}
		for i := 0; i < int(counter.Load()); i++ {
			wg.Done()
		}
		t.Error("SendSessionKeys test failed")
		os.Exit(1)
	}()
	for i := 0; i < 10; i++ {
		log.D.Ln("buying sessions", i)
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
		for j := range clients[0].SessionCache {
			log.D.F("%d %s %v", i, j, clients[0].SessionCache[j])
		}
	}
	for _, v := range clients {
		v.Shutdown()
	}
}

func TestClient_ExitTxFailureDiagnostics(t *testing.T) {
	log2.SetLogLevel(log2.Info)
	var clients []*Engine
	var e error
	if clients, e = CreateNMockCircuits(false, 1); check(e) {
		t.Error(e)
		t.FailNow()
	}
	cl := clients[0]
	// peers := clients[1:]
	const port = 3455
	sim := transport.NewSim(0)
	for i := range clients {
		if i == 0 {
			continue
		}
		_ = clients[i].AddServiceToLocalNode(&service.Service{
			Port:      port,
			Transport: sim,
			RelayRate: 18000 * 4,
		})
	}
	// Start up the clients.
	for _, v := range clients {
		go v.Start()
	}
	var wg sync.WaitGroup
	var counter atomic.Int32
	go func() {
		select {
		case <-time.After(time.Second * 5):
		}
		for i := 0; i < int(counter.Load()); i++ {
			wg.Done()
		}
		t.Error("TxFailureDiagnostics test failed")
		os.Exit(1)
	}()
	for i := 0; i < 5; i++ {
		log.D.Ln("buying sessions", i)
		wg.Add(1)
		counter.Inc()
		e = cl.BuyNewSessions(1000000, func() {
			wg.Done()
			counter.Dec()
		})
		if check(e) {
			wg.Done()
			counter.Dec()
		}
		wg.Wait()
	}
	// log.I.Ln("starting fail test")
	// log2.SetLogLevel(log2.Debug)
	// // Now we will disable each of the nodes one by one and run a discovery
	// // process to find the "failed" node.
	// for _, v := range peers {
	// 	// Pause the node (also clearing its channels and caches).
	// 	v.Pause.Signal()
	// 	// Try to send out an Exit message.
	// 	var msg slice.Bytes
	// 	if msg, _, e = tests.GenMessage(64, "request"); check(e) {
	// 		t.Error(e)
	// 		t.FailNow()
	// 	}
	// 	var respMsg slice.Bytes
	// 	var respHash sha256.Hash
	// 	if respMsg, respHash, e = tests.GenMessage(32, "response"); check(e) {
	// 		t.Error(e)
	// 		t.FailNow()
	// 	}
	// 	_ = respHash
	// 	var c traffic.Circuit
	// 	c[2] = cl.SessionCache[v.GetLocalNode().ID][2]
	// 	id := nonce.NewID()
	// 	log.D.Ln("sending out onion that will fail")
	// 	cl.SendExit(port, msg, id, cl.Sessions[1],
	// 		func(idd nonce.ID, b slice.Bytes) {
	// 			log.D.Ln("this shouldn't print!")
	// 		})
	// 	log.D.Ln("listening for message")
	// 	bb := <-clients[3].ReceiveToLocalNode(port)
	// 	log.T.S(bb.ToBytes())
	// 	if e = clients[3].SendFromLocalNode(port, respMsg); check(e) {
	// 		t.Error("fail send")
	// 	}
	// 	log.T.Ln("response sent")
	// 	// Resume the node.
	// 	v.Pause.Signal()
	// }
	for _, v := range clients {
		v.Shutdown()
	}
}

func TestClient_SendExit(t *testing.T) {
	log2.SetLogLevel(log2.Debug)
	var clients []*Engine
	var e error
	if clients, e = CreateNMockCircuits(true, 2); check(e) {
		t.Error(e)
		t.FailNow()
	}
	// set up forwarding port service
	const port = 3455
	sim := transport.NewSim(0)
	for i := range clients {
		if i == 0 {
			continue
		}
		_ = clients[i].AddServiceToLocalNode(&service.Service{
			Port:      port,
			Transport: sim,
			RelayRate: 18000 * 4,
		})
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
		t.Error("SendExit test failed")
	}()
out:
	for i := 1; i < len(clients[0].Sessions)-1; i++ {
		wg.Add(1)
		var c traffic.Circuit
		var msg slice.Bytes
		if msg, _, e = tests.GenMessage(64, "request"); check(e) {
			t.Error(e)
			t.FailNow()
		}
		var respMsg slice.Bytes
		var respHash sha256.Hash
		if respMsg, respHash, e = tests.GenMessage(32, "response"); check(e) {
			t.Error(e)
			t.FailNow()
		}
		sess := clients[0].Sessions[i]
		c[sess.Hop] = clients[0].Sessions[i]
		id := nonce.NewID()
		clients[0].SendExit(port, msg, id, clients[0].Sessions[i],
			func(idd nonce.ID, b slice.Bytes) {
				if sha256.Single(b) != respHash {
					t.Error("failed to receive expected message")
				}
				if id != idd {
					t.Error("failed to receive expected message ID")
				}
				log.I.F("success\n\n")
				wg.Done()
			})
		bb := <-clients[3].ReceiveToLocalNode(port)
		log.T.S(bb.ToBytes())
		if e = clients[3].SendFromLocalNode(port, respMsg); check(e) {
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
	for _, v := range clients {
		v.Shutdown()
	}
}

func TestClient_SendPing(t *testing.T) {
	log2.SetLogLevel(log2.Debug)
	var clients []*Engine
	var e error
	if clients, e = CreateNMockCircuits(true, 1); check(e) {
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
	for i := 1; i < len(clients[0].Sessions)-1; i++ {
		wg.Add(1)
		var c traffic.Circuit
		sess := clients[0].Sessions[i]
		c[sess.Hop] = clients[0].Sessions[i]
		clients[0].SendPing(c,
			func(id nonce.ID, b slice.Bytes) {
				log.I.Ln("success")
				wg.Done()
			})
		select {
		case <-quit:
			break out
		default:
		}
		wg.Wait()
	}
	quit.Q()
	for _, v := range clients {
		v.Shutdown()
	}
}

func TestClient_SendGetBalance(t *testing.T) {
	log2.SetLogLevel(log2.Debug)
	var clients []*Engine
	var e error
	if clients, e = CreateNMockCircuits(true, 2); check(e) {
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
		t.Error("SendGetBalance test failed")
	}()
out:
	for i := 1; i < len(clients[0].Sessions)-1; i++ {
		wg.Add(1)
		clients[0].SendGetBalance(clients[0].Sessions[i],
			func(cf nonce.ID, b slice.Bytes) {
				wg.Done()
			})
		select {
		case <-quit:
			break out
		default:
		}
		wg.Wait()
	}
	quit.Q()
	for _, v := range clients {
		v.Shutdown()
	}
}
