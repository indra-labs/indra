package engine

import (
	"context"
	
	"git-indra.lan/indra-labs/indra"
	"git-indra.lan/indra-labs/indra/pkg/crypto"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/engine/node"
	"git-indra.lan/indra-labs/indra/pkg/engine/sessions"
	"git-indra.lan/indra-labs/indra/pkg/engine/tpt"
	"git-indra.lan/indra-labs/indra/pkg/engine/transport"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	fails = log.E.Chk
)

func createNMockCircuits(inclSessions bool, nCircuits int,
	nReturnSessions int, ctx context.Context) (cl []*Engine, e error) {
	
	nTotal := 1 + nCircuits*5
	cl = make([]*Engine, nTotal)
	nodes := make([]*node.Node, nTotal)
	tpts := make([]tpt.Transport, nTotal)
	ss := make(sessions.Sessions, nTotal-1)
	for i := range tpts {
		tpts[i] = transport.NewSimDuplex(nTotal, ctx)
	}
	for i := range nodes {
		var idPrv *crypto.Prv
		if idPrv, e = crypto.GeneratePrvKey(); fails(e) {
			return
		}
		addr := slice.GenerateRandomAddrPortIPv4()
		nodes[i], _ = node.NewNode(addr, idPrv, tpts[i], 50000)
		if cl[i], e = NewEngine(Params{
			tpts[i],
			idPrv,
			nodes[i],
			nil,
			nReturnSessions},
		); fails(e) {
			return
		}
		cl[i].Manager.SetLocalNodeAddress(nodes[i].AddrPort)
		cl[i].Manager.SetLocalNode(nodes[i])
		if inclSessions {
			// Create a session for all but the first.
			if i > 0 {
				ss[i-1] = sessions.NewSessionData(nonce.NewID(), nodes[i],
					1<<16, nil, nil, byte((i-1)/nCircuits))
				// AddIntro session to node, so it will be able to relay if it
				// gets a message with the key.
				cl[i].Manager.AddSession(ss[i-1])
				// we need a copy for the node so the balance adjustments don't
				// double up.
				s := *ss[i-1]
				cl[0].Manager.AddSession(&s)
			}
		}
	}
	// Add all the nodes to each other, so they can pass messages.
	for i := range cl {
		for j := range nodes {
			if i == j {
				continue
			}
			cl[i].Manager.AddNodes(nodes[j])
		}
	}
	return
}

func CreateNMockCircuits(nCirc int, nReturns int,
	ctx context.Context) (cl []*Engine, e error) {
	
	return createNMockCircuits(false, nCirc, nReturns, ctx)
}

func CreateNMockCircuitsWithSessions(nCirc int, nReturns int,
	ctx context.Context) (cl []*Engine, e error) {
	
	return createNMockCircuits(true, nCirc, nReturns, ctx)
}
