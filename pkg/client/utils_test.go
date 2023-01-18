package client

import (
	"math"

	"github.com/indra-labs/indra/pkg/crypto/key/prv"
	"github.com/indra-labs/indra/pkg/crypto/key/pub"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/node"
	"github.com/indra-labs/indra/pkg/traffic"
	"github.com/indra-labs/indra/pkg/transport"
	"github.com/indra-labs/indra/pkg/types"
	"github.com/indra-labs/indra/pkg/util/slice"
)

func CreateNMockCircuits(inclSessions bool,
	nCircuits int) (cl []*Client, e error) {

	nTotal := 1 + nCircuits*5
	cl = make([]*Client, nTotal)
	nodes := make([]*node.Node, nTotal)
	transports := make([]types.Transport, nTotal)
	sessions := make(traffic.Sessions, nTotal-1)
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
		nodes[i], _ = node.New(addr, idPub, idPrv, transports[i])
		if cl[i], e = NewClient(transports[i], idPrv, nodes[i],
			nil); check(e) {
			return
		}
		cl[i].AddrPort = nodes[i].AddrPort
		cl[i].Node = nodes[i]
		if inclSessions {
			// create a session for all but the first
			if i > 0 {
				sessions[i-1] = traffic.NewSession(
					nonce.NewID(), nodes[i].Peer,
					math.MaxUint64, nil, nil,
					byte((i-1)/nCircuits))
				// log.I.S(sessions[i-1].Hop)
				// Add session to node, so it will be able to
				// relay if it gets a message with the key.
				nodes[i].AddSession(sessions[i-1])
				nodes[0].AddSession(sessions[i-1])
			}
		}
	}
	// for i := range nodes[0].Sessions {
	// 	log.T.S(i, nodes[0].Sessions[i].Hop)
	// }
	for i := range cl {
		for j := range nodes {
			if i == j {
				continue
			}
			cl[i].Nodes = append(cl[i].Nodes, nodes[j])
		}
	}
	return
}
