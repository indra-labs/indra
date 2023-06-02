package engine

import (
	"context"
	"github.com/indra-labs/indra"
	"sync"
	"testing"
	"time"

	"go.uber.org/atomic"

	"github.com/indra-labs/indra/pkg/util/qu"

	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/crypto/sha256"
	"github.com/indra-labs/indra/pkg/engine/services"
	"github.com/indra-labs/indra/pkg/engine/sessions"
	"github.com/indra-labs/indra/pkg/engine/transport"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/cryptorand"
	"github.com/indra-labs/indra/pkg/util/slice"
	"github.com/indra-labs/indra/pkg/util/tests"
)

func TestClient_SendExit(t *testing.T) {
	if indra.CI!="false" {
		log2.SetLogLevel(log2.Debug)
	}
	var clients []*Engine
	var e error
	ctx, cancel := context.WithCancel(context.Background())
	if clients, e = CreateNMockCircuitsWithSessions(2, 2,
		ctx); fails(e) {
		t.Error(e)
		t.FailNow()
	}
	client := clients[0]
	log.D.Ln("client", client.Manager.GetLocalNodeAddressString())
	// set up forwarding port service
	const port = 3455
	sim := transport.NewByteChan(0)
	for i := range clients {
		e = clients[i].Manager.AddServiceToLocalNode(&services.Service{
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
	for i := 3; i < len(clients[0].Manager.Sessions)-1; i++ {
		wg.Add(1)
		var msg slice.Bytes
		if msg, _, e = tests.GenMessage(64, "request"); fails(e) {
			t.Error(e)
			t.FailNow()
		}
		var respMsg slice.Bytes
		var respHash sha256.Hash
		if respMsg, respHash, e = tests.GenMessage(32,
			"response"); fails(e) {
			t.Error(e)
			t.FailNow()
		}
		bob := clients[0].Manager.Sessions[i]
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
				t.Error("failed to receive expected message Keys")
			}
			log.D.F("success\n\n")
			wg.Done()
			return
		})
		bb := <-clients[3].Mgr().GetLocalNode().ReceiveFrom(port)
		log.T.S(bb.ToBytes())
		if e = clients[3].Manager.SendFromLocalNode(port, respMsg); fails(e) {
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

func TestClient_SendPing(t *testing.T) {
	if indra.CI!="false" {
		log2.SetLogLevel(log2.Debug)
	}
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
	for i := 3; i < len(clients[0].Manager.Sessions)-1; i++ {
		wg.Add(1)
		var c sessions.Circuit
		sess := clients[0].Manager.Sessions[i]
		c[sess.Hop] = clients[0].Manager.Sessions[i]
		clients[0].SendPing(c,
			func(id nonce.ID, ifc interface{}, b slice.Bytes) (e error) {
				log.D.Ln("success")
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

func TestClient_SendSessionKeys(t *testing.T) {
	if indra.CI!="false" {
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
		for j := range clients[0].Manager.SessionCache {
			log.D.F("%d %s %v", i, j, clients[0].Manager.SessionCache[j])
		}
		quit.Q()
	}
	for _, v := range clients {
		v.Shutdown()
	}
	cancel()
}

