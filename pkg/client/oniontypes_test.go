package client

import (
	"testing"
	"time"

	"github.com/cybriq/qu"
	"github.com/indra-labs/indra/pkg/key/signer"
	"github.com/indra-labs/indra/pkg/node"
	"github.com/indra-labs/indra/pkg/nonce"
	"github.com/indra-labs/indra/pkg/session"
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
	if clients, e = CreateMockCircuitClients(nTotal); check(e) {
		t.Error(e)
		t.FailNow()
	}
	// Start up the clients.
	for _, v := range clients {
		go v.Start()
	}
	conf := nonce.NewID()
	var ks *signer.KeySet
	if _, ks, e = signer.New(); check(e) {
		t.Error(e)
		t.FailNow()
	}
	var sessions session.Sessions
	for _, v := range clients[1:] {
		sessions = append(sessions, v.Sessions[0])
	}
	sessions = append(sessions, clients[0].Sessions[0])
	os := wire.Ping(conf, sessions, ks)
	quit := qu.T()
	log.I.S("sending ping with ID", os[len(os)-1].(*confirm.OnionSkin))
	clients[0].RegisterConfirmation(func(cf nonce.ID) {
		log.I.S("received ping confirmation ID", cf)
		quit.Q()
	}, os[len(os)-1].(*confirm.OnionSkin).ID)
	o := os.Assemble()
	b := wire.EncodeOnion(o)
	clients[0].Send(clients[1].AddrPort, b)
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

func TestSendExit(t *testing.T) {
	const nTotal = 6
	clients := make([]*Client, nTotal)
	var e error
	if clients, e = CreateMockCircuitClients(nTotal); check(e) {
		t.Error(e)
		t.FailNow()
	}
	var ks *signer.KeySet
	if _, ks, e = signer.New(); check(e) {
		t.Error(e)
		t.FailNow()
	}
	var sess [3]*session.Session
	sess[0] = clients[4].Sessions.Find(clients[4].ID)
	sess[1] = clients[5].Sessions.Find(clients[5].ID)
	sess[2] = clients[0].Sessions.Find(clients[0].ID)
	clients[4].Sessions = clients[4].Sessions.Add(sess[0])
	clients[5].Sessions = clients[5].Sessions.Add(sess[1])
	clients[0].Sessions = clients[0].Sessions.Add(sess[2])
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
	var hop [nTotal - 1]*node.Node
	for i := range clients[0].Nodes {
		hop[i] = clients[0].Nodes[i]
	}
	// id := nonce.NewID()
	var message slice.Bytes
	var hash sha256.Hash
	if message, hash, e = testutils.GenerateTestMessage(32); check(e) {
		t.Error(e)
		t.FailNow()
	}
	quit := qu.T()
	os := wire.SendExit(message, port, clients[0].Node, hop, sess, ks)
	clients[0].ExitHooks = clients[0].ExitHooks.Add(hash, func() {
		log.I.S("finished")
		quit.Q()
	})
	o := os.Assemble()
	b := wire.EncodeOnion(o)
	hop[0].Send(b)
	go func() {
		time.Sleep(time.Second * 6)
		quit.Q()
		t.Error("SendExit got stuck")
	}()
	bb := <-clients[3].Services[0].Receive()
	log.I.S(bb.ToBytes())
	if e = clients[3].SendTo(port, bb); check(e) {
		t.Error("fail send")
	}
	log.I.S("response sent")
	<-quit.Wait()
	for _, v := range clients {
		v.Shutdown()
	}
}
