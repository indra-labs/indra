package client

import (
	"crypto/rand"
	"net"
	"testing"

	"github.com/Indra-Labs/indra/pkg/ifc"
	"github.com/Indra-Labs/indra/pkg/key/address"
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
	for i := 0; i < nNodes; i++ {
		ip := make(net.IP, net.IPv4len)
		if n, e = rand.Read(ip); check(e) || n != net.IPv4len {
			t.Error(e)
			t.FailNow()
		}
		var id nonce.ID
		var nod *node.Node
		nod, id = node.New(ip, transport.NewSim(0))
		nodes = append(nodes, nod)
		ids = append(ids, id)
	}
	cl := New(transport.NewSim(0), nodes)
	cl.Nodes = nodes
	var ci *onion.Circuit
	if ci, e = cl.GenerateCircuit(); check(e) {
		t.Error(e)
	}
	// Create the onion
	var lastMsg ifc.Message
	lastMsg, _, e = testutils.GenerateTestMessage(32)
	for i := range ci.Hops {
		// progress through the hops in reverse
		rm := &wire.ReturnMessage{
			IP:      ci.Hops[len(ci.Hops)-i].IP,
			Message: lastMsg,
		}
		rmm := rm.Serialize()
		ep := packet.EP{
			To: address.FromPubKey(ci.Hops[i].Key),
			// From:   _,
			Parity: 0,
			Seq:    0,
			Length: len(rmm),
			Data:   rmm,
		}
		lastMsg, e = packet.Encode(ep)
	}
	// now unwrap the message

}
