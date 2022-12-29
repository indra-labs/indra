package client

import (
	"testing"
	"time"

	"github.com/Indra-Labs/indra/pkg/ifc"
	"github.com/Indra-Labs/indra/pkg/key/prv"
	"github.com/Indra-Labs/indra/pkg/key/pub"
	"github.com/Indra-Labs/indra/pkg/key/signer"
	"github.com/Indra-Labs/indra/pkg/node"
	"github.com/Indra-Labs/indra/pkg/nonce"
	"github.com/Indra-Labs/indra/pkg/slice"
	"github.com/Indra-Labs/indra/pkg/transport"
	"github.com/Indra-Labs/indra/pkg/wire"
	"github.com/Indra-Labs/indra/pkg/wire/confirm"
	log2 "github.com/cybriq/proc/pkg/log"
	"github.com/cybriq/qu"
)

func TestPing(t *testing.T) {
	log2.CodeLoc = true
	// log2.SetLogLevel(log2.Trace)
	var clients [4]*Client
	var nodes [4]*node.Node
	var transports [4]ifc.Transport
	var e error
	for i := range transports {
		transports[i] = transport.NewSim(4)
	}
	for i := range nodes {
		var hdrPrv, pldPrv *prv.Key
		if hdrPrv, e = prv.GenerateKey(); check(e) {
			t.Error(e)
			t.FailNow()
		}
		hdrPub := pub.Derive(hdrPrv)
		if pldPrv, e = prv.GenerateKey(); check(e) {
			t.Error(e)
			t.FailNow()
		}
		pldPub := pub.Derive(pldPrv)
		addr := slice.GenerateRandomAddrPortIPv4()
		nodes[i], _ = node.New(addr, hdrPub, pldPub, hdrPrv, pldPrv, transports[i])
		if clients[i], e = New(transports[i], hdrPrv, nodes[i], nil); check(e) {
			t.Error(e)
			t.FailNow()
		}
		clients[i].AddrPort = nodes[i].AddrPort
	}
	// add each node to each other's Nodes except itself.
	for i := range nodes {
		for j := range nodes {
			if i == j {
				continue
			}
			clients[i].Nodes = append(clients[i].Nodes, nodes[j])
		}
	}
	// Start up the clients.
	for _, v := range clients {
		go v.Start()
	}
	pn := nonce.NewID()
	var ks *signer.KeySet
	if _, ks, e = signer.New(); check(e) {
		t.Error(e)
		t.FailNow()
	}
	var hop [3]*node.Node
	for i := range clients[0].Nodes {
		hop[i] = clients[0].Nodes[i]
	}
	os := wire.Ping(pn, clients[0].Node, hop, ks)
	// log.I.S(os)
	quit := qu.T()
	log.I.S("sending ping with ID", os[len(os)-1].(*confirm.OnionSkin))
	clients[0].RegisterConfirmation(
		os[len(os)-1].(*confirm.OnionSkin).ID,
		func(cf *confirm.OnionSkin) {
			log.I.S("received ping confirmation ID", cf)
			quit.Q()
		},
	)
	o := os.Assemble()
	b := wire.EncodeOnion(o)
	hop[0].Send(b)
	go func() {
		time.Sleep(time.Second * 2)
		quit.Q()
		t.Error("ping got stuck")
	}()
	<-quit.Wait()
	for _, v := range clients {
		v.Shutdown()
	}

}

//

// func TestClient_GenerateCircuit(t *testing.T) {
// 	var nodes node.Nodes
// 	var ids []nonce.ID
// 	var e error
// 	var n int
// 	nNodes := 10
// 	// Generate node private keys, session keys and keysets
// 	var prvs []*prv.Key
// 	var rcvrs []*address.Receiver
// 	var sessKeys []*prv.Key
// 	var keysets []*signer.KeySet
// 	for i := 0; i < nNodes; i++ {
// 		var p *prv.Key
// 		if p, e = prv.GenerateKey(); check(e) {
// 			t.Error(e)
// 			t.FailNow()
// 		}
// 		prvs = append(prvs, p)
// 		rcvrs = append(rcvrs, address.NewReceiver(p))
// 		var s *prv.Key
// 		var ks *signer.KeySet
// 		if s, ks, e = signer.New(); check(e) {
// 			t.Error(e)
// 			t.FailNow()
// 		}
// 		keysets = append(keysets, ks)
// 		sessKeys = append(sessKeys, s)
// 	}
// 	// create nodes using node private keys
// 	for i := 0; i < nNodes; i++ {
// 		ip := make(net.IP, net.IPv4len)
// 		if n, e = rand.Read(ip); check(e) || n != net.IPv4len {
// 			t.Error(e)
// 			t.FailNow()
// 		}
// 		var id nonce.ID
// 		var nod *node.Node
// 		nod, id = node.New(ip, pub.Derive(prvs[i]), transport.NewSim(0))
// 		nodes = append(nodes, nod)
// 		ids = append(ids, id)
// 	}
// 	var cl *Client
// 	cl, e = New(transport.NewSim(0), nodes)
// 	cl.Nodes = nodes
// 	// generate test sessions with basically infinite bandwidth
// 	for i := range cl.Nodes {
// 		sess := NewSession(cl.Nodes[i].ID,
// 			math.MaxUint64,
// 			address.NewSendEntry(cl.Nodes[i].Key),
// 			address.NewReceiveEntry(sessKeys[i]),
// 			keysets[i])
// 		cl.Sessions = cl.Sessions.Add(sess)
// 	}
// 	var ci *Circuit
// 	if ci, e = cl.GenerateReturn(); check(e) {
// 		t.Error(e)
// 		t.FailNow()
// 	}
// 	// Create the onion
// 	var lastMsg ifc.OnionSkin
// 	lastMsg, _, e = testutils.GenerateTestMessage(32)
// 	original := make([]byte, 32)
// 	copy(original, lastMsg)
// 	// log.I.S(lastMsg)
// 	// log.I.Ln(len(ci.Hops))
// 	for i := range ci.Hops {
// 		// progress through the hops in reverse
// 		rm := &wire.HeaderPub{
// 			IP:      ci.Hops[len(ci.Hops)-i-1].IP,
// 			OnionSkin: lastMsg,
// 		}
// 		rmm := rm.Encode()
// 		ep := message.EP{
// 			To: address.
// 				FromPubKey(ci.Hops[len(ci.Hops)-i-1].Key),
// 			From:   cl.Sessions[i].KeySet.Next(),
// 			Length: len(rmm),
// 			Data:   rmm,
// 		}
// 		lastMsg, e = message.Encode(ep)
// 		var to address.Cloaked
// 		var from *pub.Key
// 		if to, from, e = message.GetKeys(lastMsg); check(e) {
// 			t.Error(e)
// 			t.FailNow()
// 		}
// 		_, _ = to, from
// 		// log.I.S("lastMsg", lastMsg)
// 	}
// 	// now unwrap the message
// 	for c := 0; c < ReturnLen; c++ {
//
// 		var to address.Cloaked
// 		var from *pub.Key
// 		// log.I.S("unwrapping", c, lastMsg)
// 		if to, from, e = message.GetKeys(lastMsg); check(e) {
// 			t.Error(e)
// 			t.FailNow()
// 		}
//
// 		// log.I.S(to, from)
// 		var match *address.Receiver
// 		for i := range rcvrs {
// 			if rcvrs[i].Match(to) {
// 				match = rcvrs[i]
// 				// log.I.S(rcvrs[i].Pub)
// 				hop := rcvrs[i].Pub
// 				cct := cl.Circuits[0].Hops
// 				for j := range cct {
// 					if cct[j].Key.Equals(hop) {
// 						// log.I.Ln("found hop", j)
// 						// log.I.Ln(cct[j].IP)
// 						if j != c {
// 							t.Error("did not find expected hop")
// 							t.FailNow()
// 						}
// 						break
// 					}
// 				}
// 				break
// 			}
// 		}
// 		if match == nil {
// 			log.I.Ln("did not find matching address.Receiver")
// 			t.FailNow()
// 		}
// 		var f *message.OnionSkin
// 		if f, e = message.Decode(lastMsg, from,
// 			match.Key); check(e) {
//
// 			t.Error(e)
// 			t.FailNow()
// 		}
// 		var rm *wire.HeaderPub
// 		var msg wire.OnionSkin
// 		if msg, e = wire.Deserialize(f.Data); check(e) {
// 			t.Error(e)
// 			t.FailNow()
// 		}
// 		if rm, e = wire.ToForward(msg); check(e) {
// 			t.Error(e)
// 			t.FailNow()
// 		}
// 		// log.I.Ln(rm.IP)
// 		// log.I.S(rm.OnionSkin)
// 		// log.I.Ln(lastMsg[0], net.IP(lastMsg[1:5]))
// 		lastMsg = rm.OnionSkin
// 	}
// 	if string(original) != string(lastMsg) {
// 		t.Error("failed to recover original message")
// 		t.FailNow()
// 	}
// 	// log.I.S(lastMsg)
// }
