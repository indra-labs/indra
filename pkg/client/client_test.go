package client

import (
	"crypto/rand"
	"math"
	"net"
	"testing"

	"github.com/Indra-Labs/indra/pkg/ifc"
	"github.com/Indra-Labs/indra/pkg/key/address"
	"github.com/Indra-Labs/indra/pkg/key/prv"
	"github.com/Indra-Labs/indra/pkg/key/pub"
	"github.com/Indra-Labs/indra/pkg/key/signer"
	"github.com/Indra-Labs/indra/pkg/node"
	"github.com/Indra-Labs/indra/pkg/nonce"
	"github.com/Indra-Labs/indra/pkg/onion"
	"github.com/Indra-Labs/indra/pkg/packet"
	"github.com/Indra-Labs/indra/pkg/testutils"
	"github.com/Indra-Labs/indra/pkg/transport"
	"github.com/Indra-Labs/indra/pkg/wire"
)

func TestClient_GenerateCircuit(t *testing.T) {
	var nodes node.Nodes
	var ids []nonce.ID
	var e error
	var n int
	nNodes := 10
	// Generate node private keys, session keys and keysets
	var prvs []*prv.Key
	var sessKeys []*prv.Key
	var keysets []*signer.KeySet
	for i := 0; i < nNodes; i++ {
		var p *prv.Key
		if p, e = prv.GenerateKey(); check(e) {
			t.Error(e)
			t.FailNow()
		}
		prvs = append(prvs, p)
		var s *prv.Key
		var ks *signer.KeySet
		if s, ks, e = signer.New(); check(e) {
			t.Error(e)
			t.FailNow()
		}
		keysets = append(keysets, ks)
		sessKeys = append(sessKeys, s)
	}
	// create nodes using node private keys
	for i := 0; i < nNodes; i++ {
		ip := make(net.IP, net.IPv4len)
		if n, e = rand.Read(ip); check(e) || n != net.IPv4len {
			t.Error(e)
			t.FailNow()
		}
		var id nonce.ID
		var nod *node.Node
		nod, id = node.New(ip, pub.Derive(prvs[i]), transport.NewSim(0))
		nodes = append(nodes, nod)
		ids = append(ids, id)
	}
	var cl *Client
	cl, e = New(transport.NewSim(0), nodes)
	cl.Nodes = nodes
	// generate test sessions with basically infinite bandwidth
	for i := range cl.Nodes {
		sess := NewSession(cl.Nodes[i].ID,
			math.MaxUint64,
			address.NewSendEntry(cl.Nodes[i].Key),
			address.NewReceiveEntry(sessKeys[i]),
			keysets[i])
		cl.Sessions = cl.Sessions.Add(sess)
	}
	var ci *onion.Circuit
	if ci, e = cl.GenerateCircuit(); check(e) {
		t.Error(e)
		t.FailNow()
	}
	// Create the onion
	var lastMsg ifc.Message
	lastMsg, _, e = testutils.GenerateTestMessage(32)
	for i := range ci.Hops {
		// progress through the hops in reverse
		rm := &wire.ReturnMessage{
			IP:      ci.Hops[len(ci.Hops)-i-1].IP,
			Message: lastMsg,
		}
		rmm := rm.Serialize()
		ep := packet.EP{
			To:     address.FromPubKey(ci.Hops[i].Key),
			From:   cl.Sessions[i].KeyRoller.Next(),
			Parity: 0,
			Seq:    0,
			Length: len(rmm),
			Data:   rmm,
		}
		lastMsg, e = packet.Encode(ep)
	}
	// now unwrap the message

}
