package engine

import (
	"context"
	"math/rand"
	"sync"
	"testing"
	"time"
	
	"github.com/cybriq/qu"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/util/cryptorand"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
	"git-indra.lan/indra-labs/indra/pkg/util/tests"
)

func TestOnionSkins_Exit(t *testing.T) {
	var e error
	prvs, pubs := crypto.GetCipherSet(t)
	ciphers := crypto.GenCiphers(prvs, pubs)
	var msg slice.Bytes
	var hash sha256.Hash
	if msg, hash, e = tests.GenMessage(512, "aoeu"); fails(e) {
		t.Error(e)
		t.FailNow()
	}
	n3 := crypto.Gen3Nonces()
	p := uint16(rand.Uint32())
	id := nonce.NewID()
	ep := &ExitPoint{
		Routing: &Routing{
			Sessions: [3]*SessionData{},
			Keys:     prvs,
			Nonces:   n3,
		},
		ReturnPubs: pubs,
	}
	on := Skins{}.
		Exit(id, p, msg, ep).
		Assemble()
	s := Encode(on)
	s.SetCursor(0)
	var onc Codec
	if onc = Recognise(s); onc == nil {
		t.Error("did not unwrap")
		t.FailNow()
	}
	if e = onc.Decode(s); fails(e) {
		t.Error("did not decode")
		t.FailNow()
	}
	var ex *Exit
	var ok bool
	if ex, ok = onc.(*Exit); !ok {
		t.Error("did not unwrap expected type")
		t.FailNow()
	}
	if ex.ID != id {
		t.Error("ID did not decode correctly")
		t.FailNow()
	}
	for i := range ex.Ciphers {
		if ex.Ciphers[i] != ciphers[i] {
			t.Errorf("cipher %d did not unwrap correctly", i)
			t.FailNow()
		}
	}
	for i := range ex.Nonces {
		if ex.Nonces[i] != n3[i] {
			t.Errorf("nonce %d did not unwrap correctly", i)
			t.FailNow()
		}
	}
	plH := sha256.Single(ex.Bytes)
	if plH != hash {
		t.Errorf("exit message did not unwrap correctly")
		t.FailNow()
	}
}

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
	sim := NewByteChan(0)
	for i := range clients {
		e = clients[i].AddServiceToLocalNode(&Service{
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
		returnHops := client.SessionManager.GetSessionsAtHop(5)
		var alice *SessionData
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
