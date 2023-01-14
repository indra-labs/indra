package client

import (
	"github.com/indra-labs/indra/pkg/ifc"
	"github.com/indra-labs/indra/pkg/key/prv"
	"github.com/indra-labs/indra/pkg/key/pub"
	"github.com/indra-labs/indra/pkg/node"
	"github.com/indra-labs/indra/pkg/slice"
	"github.com/indra-labs/indra/pkg/transport"
)

func CreateMockCircuitClients(nTotal int) (clients []*Client, e error) {
	clients = make([]*Client, nTotal)
	nodes := make([]*node.Node, nTotal)
	transports := make([]ifc.Transport, nTotal)
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
	}
	// add each node to each other's Nodes except itself.
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
