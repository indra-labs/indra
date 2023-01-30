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

const DefaultTimeout = time.Second

func TestClient_SendKeys(t *testing.T) {
	log2.SetLogLevel(log2.Trace)
	var clients []*Engine
	var e error
	if clients, e = CreateNMockCircuits(false, 2, DefaultTimeout); check(e) {
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
		t.Error("SendExit test failed")
		os.Exit(1)
	}()

	wg.Wait()
	for _, v := range clients {
		v.Shutdown()
	}
}

func TestClient_SendPing(t *testing.T) {
	log2.SetLogLevel(log2.Debug)
	var clients []*Engine
	var e error
	if clients, e = CreateNMockCircuits(true, 2, DefaultTimeout); check(e) {
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

func TestClient_SendExit(t *testing.T) {
	log2.SetLogLevel(log2.Debug)
	var clients []*Engine
	var e error
	if clients, e = CreateNMockCircuits(true, 2, DefaultTimeout); check(e) {
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
		clients[i].Services = append(clients[i].Services, &service.Service{
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
		bb := <-clients[3].Services[0].Receive()
		log.T.S(bb.ToBytes())
		if e = clients[3].SendTo(port, respMsg); check(e) {
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

func TestClient_SendGetBalance(t *testing.T) {
	log2.SetLogLevel(log2.Debug)
	var clients []*Engine
	var e error
	if clients, e = CreateNMockCircuits(true, 2, DefaultTimeout); check(e) {
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
				log.I.F("success\n\n")
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
