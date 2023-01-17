package client

import (
	"testing"
	"time"

	"github.com/cybriq/qu"
	"github.com/indra-labs/indra/pkg/crypto/key/prv"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/crypto/sha256"
	"github.com/indra-labs/indra/pkg/onion"
	"github.com/indra-labs/indra/pkg/onion/layers/confirm"
	"github.com/indra-labs/indra/pkg/onion/layers/session"
	"github.com/indra-labs/indra/pkg/payment"
	"github.com/indra-labs/indra/pkg/service"
	"github.com/indra-labs/indra/pkg/traffic"
	"github.com/indra-labs/indra/pkg/transport"
	"github.com/indra-labs/indra/pkg/util/slice"
	"github.com/indra-labs/indra/pkg/util/tests"
)

func TestPing(t *testing.T) {
	const nTotal = 6
	clients := make([]*Client, nTotal)
	var e error
	if clients, e = CreateNMockCircuits(true, 1); check(e) {
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
		log.T.S("received ping confirmation ID", cf)
		quit.Q()
	}, os[len(os)-1].(*confirm.Layer).ID)
	b := onion.Encode(os.Assemble())
	log.T.S("sending ping with ID", os[len(os)-1].(*confirm.Layer))
	clients[0].Send(clients[0].Nodes[0].AddrPort, b)
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
	const nTotal = 6
	clients := make([]*Client, nTotal)
	var e error
	if clients, e = CreateNMockCircuits(true, 1); check(e) {
		t.Error(e)
		t.FailNow()
	}
	// set up forwarding port service
	const port = 3455
	clients[3].Services = append(clients[3].Services, &service.Service{
		Port:      port,
		Transport: transport.NewSim(0),
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
	var msgHash sha256.Hash
	if msg, msgHash, e = tests.GenMessage(32, "request"); check(e) {
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
	os := onion.SendExit(msg, port, clients[0].GetSessionByIndex(0), circuit,
		clients[0].KeySet)
	clients[0].Hooks = clients[0].Hooks.Add(msgHash,
		func(b slice.Bytes) {
			log.T.S(b.ToBytes())
			if sha256.Single(b) != respHash {
				t.Error("failed to receive expected message")
			}
			quit.Q()
		})
	b := onion.Encode(os.Assemble())
	log.T.Ln(clients[0].Node.AddrPort.String())
	clients[0].Node.Send(b)
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
	var clients []*Client
	var e error
	if clients, e = CreateNMockCircuits(false, 1); check(e) {
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
		sess[i] = session.New()
		pmt[i] = sess[i].ToPayment(1000000)
		clients[i+1].PaymentChan <- pmt[i]
	}
	// Send the keys.
	var circuit traffic.Circuit
	for i := range circuit {
		circuit[i] = traffic.NewSession(pmt[i].ID,
			clients[i+1].Node.Peer, pmt[i].Amount,
			sess[i].Header, sess[i].Payload, byte(i))
	}
	var hdr, pld [5]*prv.Key
	for i := range hdr {
		hdr[i], pld[i] = sess[i].Header, sess[i].Payload
	}
	sk := onion.SendKeys(cnf, hdr, pld, clients[0].GetSessionByIndex(0),
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
	var clients []*Client
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
	go func() {
		select {
		case <-time.After(time.Second):
			// t.Error("getbalance got stuck")
			quit.Q()
		case <-quit:
		}
	}()

	<-quit.Wait()
	for _, v := range clients {
		v.Shutdown()
	}
}
