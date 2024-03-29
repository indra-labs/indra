package engine

import (
	"context"
	"crypto/rand"
	"errors"
	"git.indra-labs.org/dev/ind/pkg/crypto"
	"git.indra-labs.org/dev/ind/pkg/crypto/nonce"
	"git.indra-labs.org/dev/ind/pkg/crypto/sha256"
	"git.indra-labs.org/dev/ind/pkg/engine/node"
	"git.indra-labs.org/dev/ind/pkg/engine/sessions"
	"git.indra-labs.org/dev/ind/pkg/engine/tpt"
	"git.indra-labs.org/dev/ind/pkg/engine/transport"
	log2 "git.indra-labs.org/dev/ind/pkg/proc/log"
	"git.indra-labs.org/dev/ind/pkg/util/slice"
	"github.com/multiformats/go-multiaddr"
	"os"
)

var (
	log   = log2.GetLogger()
	fails = log.E.Chk
)

// todo: none of this is in use anymore.

// CreateNMockCircuitsWithSessions creates an arbitrary number of mock circuits
// from the given specification, with an arbitrary number of mock sessions.
func CreateNMockCircuitsWithSessions(nCirc int, nReturns int,
	ctx context.Context) (cl []*Engine, e error) {
	return createNMockCircuits(true, nCirc, nReturns, ctx)
}

// CreateNMockCircuits creates an arbitrary number of mock circuits
// from the given specification, with an arbitrary number of mock sessions.
func CreateNMockCircuits(nCirc int, nReturns int,
	ctx context.Context) (cl []*Engine, e error) {
	return createNMockCircuits(false, nCirc, nReturns, ctx)
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
	var seeds []string
	for i := range nodes {
		var id *crypto.Keys
		if id, e = crypto.GenerateKeys(); fails(e) {
			return
		}
		addr := slice.GenerateRandomAddrPortIPv4()
		nodes[i], _ = node.NewNode([]multiaddr.Multiaddr{addr}, id, tpts[i], 50000)
		var l *transport.Listener
		var dataPath string
		dataPath, e = os.MkdirTemp(os.TempDir(), "badger")
		if e != nil {
			return
		}
		var k *crypto.Keys
		if k, e = crypto.GenerateKeys(); fails(e) {
			return
		}
		secret := sha256.New()
		rand.Read(secret[:])
		store, closer := transport.BadgerStore(dataPath, secret[:])
		if store == nil {
			return nil, errors.New("could not open database")
		}
		if l, e = transport.NewListener(seeds,
			[]string{transport.LocalhostZeroIPv4TCP,
				transport.LocalhostZeroIPv6TCP},
			k, store, closer, ctx, transport.DefaultMTU, nil); fails(e) {

			return
		}
		if i == 0 {
			seeds = transport.GetHostMultiaddrs(l.Host)
		}
		if cl[i], e = New(Params{
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
		cl[i].Mgr().AddLocalNodeAddresses(nodes[i].Addresses)
		cl[i].Mgr().SetLocalNode(nodes[i])
		if inclSessions {
			// Create a session for all but the first.
			if i > 0 {
				ss[i-1] = sessions.NewSessionData(nonce.NewID(), nodes[i],
					1<<16, nil, nil, byte((i-1)/nCircuits))
				// AddIntro session to node, so it will be able to relay if it
				// gets a message with the key.
				cl[i].Mgr().AddSession(ss[i-1])
				// we need a copy for the node so the balance adjustments don't
				// double up.
				s := *ss[i-1]
				cl[0].Mgr().AddSession(&s)
			}
		}
	}
	// Add all the nodes to each other, so they can pass messages.
	for i := range cl {
		for j := range nodes {
			if i == j {
				continue
			}
			cl[i].Mgr().AddNodes(nodes[j])
		}
	}
	return
}
