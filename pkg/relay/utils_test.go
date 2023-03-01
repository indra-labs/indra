package relay

import (
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/relay/transport"
	"git-indra.lan/indra-labs/indra/pkg/relay/types"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

func createNMockCircuits(inclSessions bool, nCircuits int,
	nReturnSessions int) (cl []*Engine, e error) {
	
	nTotal := 1 + nCircuits*5
	cl = make([]*Engine, nTotal)
	nodes := make([]*Node, nTotal)
	tpts := make([]types.Transport, nTotal)
	sessions := make(Sessions, nTotal-1)
	for i := range tpts {
		tpts[i] = transport.NewSim(nTotal)
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
		nodes[i], _ = NewNode(addr, idPub, idPrv, tpts[i], 50000, local)
		if cl[i], e = NewEngine(EngineParams{
			tpts[i],
			idPrv,
			nodes[i],
			nil,
			nReturnSessions},
		); check(e) {
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
				// AddIntro session to node, so it will be able to relay if it gets a
				// message with the key.
				cl[i].AddSession(sessions[i-1])
				// we need a copy for the node so the balance adjustments don't
				// double up.
				s := *sessions[i-1]
				cl[0].AddSession(&s)
			}
		}
	}
	// AddIntro all the nodes to each other, so they can pass messages.
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
func CreateNMockCircuits(nCirc int, nReturns int) (cl []*Engine, e error) {
	return createNMockCircuits(false, nCirc, nReturns)
}

func CreateNMockCircuitsWithSessions(nCirc int, nReturns int) (cl []*Engine,
	e error) {
	
	return createNMockCircuits(true, nCirc, nReturns)
}
