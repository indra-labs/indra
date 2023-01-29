package indra

import (
	"sync"
	"testing"
	"time"

	"github.com/cybriq/qu"

	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/node"
	"git-indra.lan/indra-labs/indra/pkg/onion"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/confirm"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/session"
	"git-indra.lan/indra-labs/indra/pkg/payment"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/service"
	"git-indra.lan/indra-labs/indra/pkg/traffic"
	"git-indra.lan/indra-labs/indra/pkg/transport"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
	"git-indra.lan/indra-labs/indra/pkg/util/tests"
)

func TestPing(t *testing.T) {
	log2.SetLogLevel(log2.Trace)
	const nTotal = 6
	clients := make([]*Engine, nTotal)
	var e error
	if clients, e = CreateNMockCircuits(true, 1, DefaultTimeout); check(e) {
		t.Error(e)
		t.FailNow()
	}
	// Start up the clients.
	for _, v := range clients {
		go v.Start()
	}
	conf := nonce.NewID()
	var circuit traffic.Circuit
	for i := range circuit {
		circuit[i] = clients[i+1].GetSessionByIndex(1)
	}
	os := onion.Ping(conf, clients[0].GetSessionByIndex(0),
		circuit, clients[0].KeySet)
	quit := qu.T()
	clients[0].RegisterConfirmation(func(cf nonce.ID) {
		if cf == conf {
			log.T.S("received ping confirmation ID", cf)
			quit.Q()
		}
	}, os[len(os)-1].(*confirm.Layer).ID)
	log.T.S("sending ping with ID", os[len(os)-1].(*confirm.Layer))
	clients[0].SendOnion(clients[0].Nodes[0].AddrPort, os, nil)
	go func() {
		select {
		case <-time.After(time.Second):
			t.Error("ping got stuck")
			quit.Q()
		case <-quit:
		}
	}()
	<-quit.Wait()
	for _, v := range clients {
		v.Shutdown()
	}
}

func TestSendExit(t *testing.T) {
	log2.SetLogLevel(log2.Trace)
	const nTotal = 6
	clients := make([]*Engine, nTotal)
	var e error
	if clients, e = CreateNMockCircuits(true, 1, DefaultTimeout); check(e) {
		t.Error(e)
		t.FailNow()
	}
	// set up forwarding port service
	const port = 3455
	clients[3].Services = append(clients[3].Services, &service.Service{
		Port:      port,
		Transport: transport.NewSim(0),
		RelayRate: 18000 * 4,
	})
	// Start up the clients.
	for _, v := range clients {
		go v.Start()
	}
	var circuit traffic.Circuit
	for i := range circuit {
		circuit[i] = clients[0].GetSessionByIndex(i + 1)
	}
	var msg slice.Bytes
	if msg, _, e = tests.GenMessage(32, "request"); check(e) {
		t.Error(e)
		t.FailNow()
	}
	var respMsg slice.Bytes
	var respHash sha256.Hash
	if respMsg, respHash, e = tests.GenMessage(32, "response"); check(e) {
		t.Error(e)
		t.FailNow()
	}
	quit := qu.T()
	id := nonce.NewID()
	os := onion.SendExit(port, msg, id, clients[0].GetSessionByIndex(0),
		circuit, clients[0].KeySet)

	hook := func(idd nonce.ID, b slice.Bytes) {
		log.T.S(b.ToBytes())
		if sha256.Single(b) != respHash {
			t.Error("failed to receive expected message")
		}
		if id != idd {
			t.Errorf("failed to receive expected message ID got %x, "+
				"received %x", idd, id)
		}
		quit.Q()
	}
	clients[0].SendOnion(clients[0].Nodes[0].AddrPort, os, hook)
	go func() {
		select {
		case <-time.After(time.Second):
			t.Error("sendexit got stuck")
			quit.Q()
		case <-quit:
		}
	}()
	bb := <-clients[3].Services[0].Receive()
	log.T.S(bb.ToBytes())
	if e = clients[3].SendTo(port, respMsg); check(e) {
		t.Error("fail send")
	}
	log.T.Ln("response sent")
	<-quit.Wait()
	for _, v := range clients {
		v.Shutdown()
	}
}

func TestSendKeys(t *testing.T) {
	log2.SetLogLevel(log2.Trace)
	var clients []*Engine
	var e error
	if clients, e = CreateNMockCircuits(false, 1, DefaultTimeout); check(e) {
		t.Error(e)
		t.FailNow()
	}
	// Start up the clients.
	for _, v := range clients {
		go v.Start()
	}
	quit := qu.T()
	go func() {
		select {
		case <-time.After(time.Second):
			t.Error("sendkeys got stuck")
			quit.Q()
		case <-quit:
		}
	}()
	cnf := nonce.NewID()
	var sess [5]*session.Layer
	var pmt [5]*payment.Payment
	for i := range clients[1:] {
		// Create a new payment and drop on the payment channel.
		sess[i] = session.New(byte(i))
		pmt[i] = sess[i].ToPayment(1000000)
		clients[i+1].Peer.PaymentChan <- pmt[i]
	}
	// Send the keys.
	circuit := make(node.Nodes, 5)
	for i := range circuit {
		circuit[i] = clients[i+1].Node
	}
	sk := onion.SendKeys(cnf, sess, clients[0].GetSessionByIndex(0),
		circuit, clients[0].KeySet)
	clients[0].RegisterConfirmation(func(cf nonce.ID) {
		log.T.S("received payment confirmation ID", cf)
		if cf != cnf {
			t.Errorf("did not receive expected confirmation, got:"+
				" %x expected: %x", cf, cnf)
			t.FailNow()
		}
		quit.Q()
	}, cnf)
	b := onion.Encode(sk.Assemble())
	clients[0].Send(clients[0].Nodes[0].AddrPort, b)
	<-quit.Wait()
	for _, v := range clients {
		v.Shutdown()
	}
}

func TestGetBalance(t *testing.T) {
	log2.SetLogLevel(log2.Trace)
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
		t.Error("test failed")
	}()
	// Sessions are generated by the function above as in sequential order
	// of session hop position such that it goes 0,0,1,1,2,2 and so on,
	// repeating the hop position N times as the number parameter in
	// CreateNMockCircuits(). So we iterate odd and even in this case and
	// then multiply the iterator by the circuit hop position to generate
	// the index for it.
	var circuit traffic.Circuit
	_ = circuit
out:
	for odd := 0; odd < 2; odd++ {
		for circ := 0; circ < 5; circ++ {
			wg.Add(1)
			// one is added to skip the return hop.
			first := odd + 1
			// fill the hops until the target in the circuit - the
			// remainder will not be accessed anyway, and last hop
			// actually we can just put the last hop, but for
			// testing this doesn't matter, just fill to the point.
			for i := 0; i < circ+1; i++ {
				circuit[circ] = clients[0].
					GetSessionByIndex(first)
				first += 2
			}
			// Now form the return hops from the last two of the
			// complementary set.
			firstR := 10 - odd - 2
			secondR := firstR + 2
			returns := [3]*traffic.Session{
				clients[0].GetSessionByIndex(firstR),
				clients[0].GetSessionByIndex(secondR),
				clients[0].GetSessionByIndex(0),
			}
			nn := nonce.NewID()
			o := onion.GetBalance(circuit, circ, returns, clients[0].KeySet,
				nn)
			clients[0].SendOnion(circuit[0].AddrPort, o, nil)
			clients[0].RegisterConfirmation(func(cf nonce.ID) {
				log.I.Ln("success")
				wg.Done()
			}, nn)
			select {
			case <-quit:
				break out
			default:
			}
			wg.Wait()
		}
		quit.Q()
	}
	for _, v := range clients {
		v.Shutdown()
	}
}
