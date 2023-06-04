package engine

import (
	"context"
	"github.com/indra-labs/indra"
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/engine/node"
	"github.com/indra-labs/indra/pkg/engine/sessions"
	"github.com/indra-labs/indra/pkg/engine/tpt"
	"github.com/indra-labs/indra/pkg/engine/transport"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/slice"
	"github.com/multiformats/go-multiaddr"
	"net/netip"
	"os"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	fails = log.E.Chk
)

func CreateNMockCircuits(nCirc int, nReturns int,
	ctx context.Context) (cl []*Engine, e error) {
	return createNMockCircuits(false, nCirc, nReturns, ctx)
}

func CreateNMockCircuitsWithSessions(nCirc int, nReturns int,
	ctx context.Context) (cl []*Engine, e error) {
	return createNMockCircuits(true, nCirc, nReturns, ctx)
}

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
		var id *crypto.Keys
		if id, e = crypto.GenerateKeys(); fails(e) {
			return
		}
		addr := slice.GenerateRandomAddrPortIPv4()
		nodes[i], _ = node.NewNode(addr, id, tpts[i], 50000)
		if cl[i], e = NewEngine(Params{
			Listener: &transport.Listener{
				MTU: 1382,
			},
			Transport:       tpts[i],
			Keys:            id,
			Node:            nodes[i],
			NReturnSessions: nReturnSessions,
		}); fails(e) {
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

// CreateMockEngine creates an indra Engine with a random localhost listener.
func CreateMockEngine(seed, dataPath string) (ng *Engine, cancel func(), e error) {
	var ctx context.Context
	ctx, cancel = context.WithCancel(context.Background())
	var keys []*crypto.Keys
	var nodes []*node.Node
	var k *crypto.Keys
	if k, e = crypto.GenerateKeys(); fails(e) {
		return
	}
	keys = append(keys, k)
	var l *transport.Listener
	if l, e = transport.NewListener(seed, transport.LocalhostZeroIPv4TCP,
		dataPath, k, ctx, transport.DefaultMTU); fails(e) {
		os.RemoveAll(dataPath)
		return
	}
	sa := transport.GetHostAddress(l.Host)
	var addr netip.AddrPort
	var ma multiaddr.Multiaddr
	if ma, e = multiaddr.NewMultiaddr(sa); fails(e) {
		e = os.RemoveAll(dataPath)
		return
	}

	var ip, port string
	if ip, e = ma.ValueForProtocol(multiaddr.P_IP4); fails(e) {
		// we specified ipv4 previously.
		fails(os.RemoveAll(dataPath))
		return
	}
	if port, e = ma.ValueForProtocol(multiaddr.P_TCP); fails(e) {
		fails(os.RemoveAll(dataPath))
		return
	}
	if addr, e = netip.ParseAddrPort(ip + ":" + port); fails(e) {
		fails(os.RemoveAll(dataPath))
		return
	}
	var nod *node.Node
	if nod, _ = node.NewNode(&addr, k, nil, 50000); fails(e) {
		fails(os.RemoveAll(dataPath))
		return
	}
	nodes = append(nodes, nod)
	if ng, e = NewEngine(Params{
		Transport: transport.NewByteChan(transport.ConnBufs),
		Listener: l,
		Keys:     k,
		Node:     nod,
	}); fails(e) {
		os.RemoveAll(dataPath)
		return
	}
	return
}
