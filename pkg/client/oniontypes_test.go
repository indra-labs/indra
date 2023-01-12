package client

import (
	"testing"
	"time"

	"github.com/cybriq/qu"
	"github.com/indra-labs/indra/pkg/key/signer"
	"github.com/indra-labs/indra/pkg/node"
	"github.com/indra-labs/indra/pkg/nonce"
	"github.com/indra-labs/indra/pkg/sha256"
	"github.com/indra-labs/indra/pkg/slice"
	"github.com/indra-labs/indra/pkg/testutils"
	"github.com/indra-labs/indra/pkg/transport"
	"github.com/indra-labs/indra/pkg/wire"
	"github.com/indra-labs/indra/pkg/wire/confirm"
)

func TestPing(t *testing.T) {
	const nTotal = 6
	clients := make([]*Client, nTotal)
	var e error
	if clients, e = CreateMockCircuitClients(); check(e) {
		t.Error(e)
		t.FailNow()
	}
	// Start up the clients.
	for _, v := range clients {
		go v.Start()
	}
	conf := nonce.NewID()
	var circuit node.Circuit
	for i := range circuit {
		circuit[i] = clients[i+1].Sessions[1]
	}
	os := wire.Ping(conf, clients[0].Node, circuit, clients[0].KeySet)
	quit := qu.T()
	clients[0].RegisterConfirmation(func(cf nonce.ID) {
		log.I.S("received ping confirmation ID", cf)
		quit.Q()
	}, os[len(os)-1].(*confirm.OnionSkin).ID)
	o := os.Assemble()
	b := wire.EncodeOnion(o)
	log.T.S("sending ping with ID", os[len(os)-1].(*confirm.OnionSkin))
	clients[0].Send(clients[0].Nodes[0].AddrPort, b)
	go func() {
		select {
		case <-time.After(time.Second):
			t.Error("ping got stuck")
		case <-quit:
		}
		time.Sleep(time.Second)
		quit.Q()
	}()
	<-quit.Wait()
	for _, v := range clients {
		v.Shutdown()
	}
}

func TestSendExit(t *testing.T) {
	const nTotal = 6
	clients := make([]*Client, nTotal)
	var e error
	if clients, e = CreateMockCircuitClients(); check(e) {
		t.Error(e)
		t.FailNow()
	}
	var ks *signer.KeySet
	if _, ks, e = signer.New(); check(e) {
		t.Error(e)
		t.FailNow()
	}
	// set up forwarding port service
	const port = 3455
	clients[3].Services = append(clients[3].Services, &node.Service{
		Port:      port,
		Transport: transport.NewSim(0),
	})
	// Start up the clients.
	for _, v := range clients {
		go v.Start()
	}
	var circuit node.Circuit
	for i := range circuit {
		circuit[i] = clients[0].Sessions[i+1]
	}
	var message slice.Bytes
	var hash sha256.Hash
	if message, hash, e = testutils.GenerateTestMessage(32, "request"); check(e) {
		t.Error(e)
		t.FailNow()
	}
	var responseMessage slice.Bytes
	var responseHash sha256.Hash
	if responseMessage, responseHash, e =
		testutils.GenerateTestMessage(32, "response"); check(e) {

		t.Error(e)
		t.FailNow()
	}
	quit := qu.T()
	os := wire.SendExit(message, port, clients[0].Node, circuit, ks)
	clients[0].ExitHooks = clients[0].ExitHooks.Add(hash,
		func(b slice.Bytes) {
			log.T.S(b.ToBytes())
			if sha256.Single(b) != responseHash {
				t.Error("failed to receive expected message")
			}
			quit.Q()
		})
	o := os.Assemble()
	b := wire.EncodeOnion(o)
	log.T.Ln(clients[0].Node.AddrPort.String())
	clients[0].Node.Send(b)
	go func() {
		time.Sleep(time.Second * 6)
		quit.Q()
		t.Error("SendExit got stuck")
	}()
	bb := <-clients[3].Services[0].Receive()
	log.T.S(bb.ToBytes())
	if e = clients[3].SendTo(port, responseMessage); check(e) {
		t.Error("fail send")
	}
	log.T.Ln("response sent")
	<-quit.Wait()
	for _, v := range clients {
		v.Shutdown()
	}
}

// func TestSendKeys(t *testing.T) {
// 	const nTotal = 6
// 	clients := make([]*Client, nTotal)
// 	var e error
// 	if clients, e = CreateMockCircuitClients(nTotal); check(e) {
// 		t.Error(e)
// 		t.FailNow()
// 	}
// 	// Start up the clients.
// 	for _, v := range clients {
// 		go v.Start()
// 	}
// 	quit := qu.T()
// 	clients[0].SendKeys(clients[0].Nodes[0].ID, func(cf nonce.ID) {
// 		log.I.S("received sendkeys confirmation ID", cf)
// 		quit.Q()
// 	})
// 	<-quit.Wait()
// 	for _, v := range clients {
// 		v.Shutdown()
// 	}
// }
