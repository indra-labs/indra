package relay

import (
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/transport"
	"git-indra.lan/indra-labs/indra/pkg/types"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

func CreateNMockCircuits(inclSessions bool, nCircuits int, nReturnSessions int) (cl []*Engine, e error) {

	nTotal := 1 + nCircuits*5
	cl = make([]*Engine, nTotal)
	nodes := make([]*Node, nTotal)
	transports := make([]types.Transport, nTotal)
	sessions := make(Sessions, nTotal-1)
	for i := range transports {
		transports[i] = transport.NewSim(nTotal)
	}
	for i := range nodes {
		var idPrv *prv.Key
		if idPrv, e = prv.GenerateKey(); check(e) {
			return
		}
		idPub := pub.Derive(idPrv)
		addr := slice.GenerateRandomAddrPortIPv4()
		var local bool
		if i == 0 {
			local = true
		}
		nodes[i], _ = NewNode(addr, idPub, idPrv, transports[i], 50000,
			local)
		if cl[i], e = NewEngine(transports[i], idPrv, nodes[i], nil,
			nReturnSessions); check(e) {
			return
		}
		cl[i].SetLocalNodeAddress(nodes[i].AddrPort)
		cl[i].SetLocalNode(nodes[i])
		if inclSessions {
			// Create a session for all but the first.
			if i > 0 {
				sessions[i-1] = NewSession(
					nonce.NewID(), nodes[i],
					1<<16, nil, nil,
					byte((i-1)/nCircuits))
				// Add session to node, so it will be able to relay if it gets a
				// message with the key.
				cl[i].AddSession(sessions[i-1])
				// we need a copy for the node so the balance adjustments don't
				// double up.
				s := *sessions[i-1]
				cl[0].AddSession(&s)
			}
		}
	}
	// Add all the nodes to each other, so they can pass messages.
	for i := range cl {
		for j := range nodes {
			if i == j {
				continue
			}
			cl[i].AddNodes(nodes[j])
		}
	}
	return
}
