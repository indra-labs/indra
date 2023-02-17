package relay

import (
	"os"
	"sync"
	"testing"
	"time"
	
	"go.uber.org/atomic"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/service"
	"git-indra.lan/indra-labs/indra/pkg/traffic"
	"git-indra.lan/indra-labs/indra/pkg/transport"
	"git-indra.lan/indra-labs/indra/pkg/util/cryptorand"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
	"git-indra.lan/indra-labs/indra/pkg/util/tests"
)

func TestClient_ExitTxFailureDiagnostics(t *testing.T) {
	log2.SetLogLevel(log2.Info)
	var clients []*Engine
	var e error
	if clients, e = CreateNMockCircuits(false, 10, 10); check(e) {
		t.Error(e)
		t.FailNow()
	}
	cl := clients[0]
	peers := clients[1:]
	const port = 3455
	sim := transport.NewSim(0)
	// Start up the clients.
	for _, v := range clients {
		go v.Start()
	}
	var wg sync.WaitGroup
	var counter atomic.Int32
	go func() {
		select {
		case <-time.After(time.Second * 20):
		}
		for i := 0; i < int(counter.Load()); i++ {
			wg.Done()
		}
		t.Error("TxFailureDiagnostics test failed")
		os.Exit(1)
	}()
	// We need 25 sessions to fill all slots.
	for i := 0; i < 50; i++ {
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
	for i := range peers {
		for j := range peers[i].Sessions {
			log.D.Ln(peers[i].GetLocalNodeAddress(),
				peers[i].Sessions[j].Hop)
		}
		e = peers[i].AddServiceToLocalNode(&service.Service{
			Port:      port,
			Transport: sim,
			RelayRate: 18000 * 4,
		})
		if check(e) {
			t.Error(e)
			t.FailNow()
		}
	}
	log.I.Ln("starting fail test")
	log2.SetLogLevel(log2.Trace)
	// Now we will disable each of the nodes one by one and run a discovery
	// process to find the "failed" node.
	for _, v := range peers {
		// Try to send out an Exit message.
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
		_ = respHash
		var c traffic.Circuit
		c[2] = cl.SessionCache[v.GetLocalNode().ID][2]
		id := nonce.NewID()
		log.D.Ln("sending out onion that will fail")
		exits := v.GetSessionsAtHop(5)
		cryptorand.Shuffle(len(exits), func(i, j int) {
			exits[i], exits[j] = exits[j], exits[i]
		})
		c, o := cl.MakeExit(port, msg, id, exits[0])
		failHop := cryptorand.IntN(5)
		nodeID := c[failHop].Node.ID
		var failClient *Engine
		for i := range peers {
			if peers[i].GetLocalNode().ID == nodeID {
				failClient = peers[i]
				break
			}
		}
		if failClient == nil {
			t.Error("did not find node to pause")
			continue
		}
		// Pause the selected client.
		failClient.Pause.Signal()
		cl.SendExitNew(c, o, func(idd nonce.ID,
			b slice.Bytes) {
			log.D.Ln("this probably won't print!")
		}, 0)
		if failHop > 2 {
			log.D.Ln("listening for message")
			bb := <-clients[3].ReceiveToLocalNode(port)
			log.T.S(bb.ToBytes())
			if e = clients[3].SendFromLocalNode(port, respMsg); check(e) {
				t.Error("fail send")
			}
			log.D.Ln("response sent")
		}
		time.Sleep(time.Second)
		// Resume the selected client.
		failClient.Pause.Signal()
	}
	for _, v := range clients {
		v.Shutdown()
	}
}
