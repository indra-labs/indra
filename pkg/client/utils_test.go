package client

import (
	"math"

	"github.com/indra-labs/indra/pkg/ifc"
	"github.com/indra-labs/indra/pkg/key/prv"
	"github.com/indra-labs/indra/pkg/key/pub"
	"github.com/indra-labs/indra/pkg/node"
	"github.com/indra-labs/indra/pkg/nonce"
	"github.com/indra-labs/indra/pkg/slice"
	"github.com/indra-labs/indra/pkg/transport"
)

const nTotal = 6

func CreateMockCircuitClients() (clients []*Client, e error) {
	clients = make([]*Client, nTotal)
	nodes := make([]*node.Node, nTotal)
	transports := make([]ifc.Transport, nTotal)
	sessions := make(node.Sessions, nTotal-1)
	for i := range transports {
		transports[i] = transport.NewSim(nTotal)
	}
	for i := range nodes {
		var hdrPrv *prv.Key
		if hdrPrv, e = prv.GenerateKey(); check(e) {
			return
		}
		hdrPub := pub.Derive(hdrPrv)
		addr := slice.GenerateRandomAddrPortIPv4()
		nodes[i], _ = node.New(addr, hdrPub, hdrPrv, transports[i])
		if clients[i], e = New(transports[i], hdrPrv, nodes[i],
			nil); check(e) {
			return
		}
		clients[i].AddrPort = nodes[i].AddrPort
		// create a session for all but the first
		if i > 0 {
			sessions[i-1] = node.NewSession(nonce.NewID(),
				nodes[i], math.MaxUint64)
			// Add session to node, so it will be able to relay if
			// it gets a message with the key.
			nodes[i].Sessions = append(nodes[i].Sessions,
				sessions[i-1])
			// Normally only the client would have this in its
			// nodes, but we are sharing them for simple circuit
			// tests. Relays don't use this field, though clients
			// can be relays.
			nodes[i].Circuit = &node.Circuit{}
			nodes[i].Circuit[i-1] = sessions[i-1]
		}
	}
	// Add each node to each other's Nodes except itself, this enables them
	// to send messages across their transports to each other, as well as,
	// in the case of our simulation, providing the circuit which was added
	// to them just before.
	for i := range clients {
		for j := range nodes {
			if i == j {
				continue
			}
			clients[i].Nodes = append(clients[i].Nodes, nodes[j])
		}
	}

	return
}
