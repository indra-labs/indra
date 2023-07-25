package engine

import (
	"context"
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/engine/node"
	"github.com/indra-labs/indra/pkg/engine/transport"
	"github.com/indra-labs/indra/pkg/util/multi"
	"github.com/multiformats/go-multiaddr"
	"net/netip"
	"os"
)

// CreateMockEngine creates an indra Engine with a random localhost listener.
func CreateMockEngine(seed, dataPath string, ctx context.Context) (ng *Engine) {
	var e error
	defer func(f *error) {
		if *f != nil {
			fails(os.RemoveAll(dataPath))
		}
	}(&e)
	var keys []*crypto.Keys
	var k *crypto.Keys
	if k, e = crypto.GenerateKeys(); fails(e) {
		return
	}
	keys = append(keys, k)
	store, closer := transport.BadgerStore(dataPath)
	if store == nil {
		log.E.Ln("could not open database")
		return nil
	}
	var l *transport.Listener
	if l, e = transport.NewListener([]string{seed},
		[]string{transport.LocalhostZeroIPv4TCP,
			transport.LocalhostZeroIPv6TCP}, k, store, closer, ctx,
		transport.DefaultMTU); fails(e) {
		return
	}
	if l == nil {
		panic("maybe you have no network device?")
	}
	sa := transport.GetHostFirstMultiaddr(l.Host)
	var ap netip.AddrPort
	var ma multiaddr.Multiaddr
	if ma, e = multiaddr.NewMultiaddr(sa); fails(e) {
		return
	}
	if ap, e = multi.AddrToAddrPort(ma); fails(e) {
		return
	}
	var nod *node.Node
	if nod, _ = node.NewNode([]*netip.AddrPort{&ap}, k, nil, 50000); fails(e) {
		return
	}
	if ng, e = New(Params{
		Transport: transport.NewDuplexByteChan(transport.ConnBufs),
		Listener:  l,
		Keys:      k,
		Node:      nod,
	}); fails(e) {
	}
	return
}

func CreateAndStartMockEngines(n int, ctx context.Context) (engines []*Engine,
	cleanup func(), e error) {

	cleanup = func() {}
	var seed string
	dataPath := make([]string, n)
	for i := 0; i < n; i++ {
		dataPath[i], e = os.MkdirTemp(os.TempDir(), "badger")
		if e != nil {
			return
		}
		var eng *Engine
		if eng = CreateMockEngine(seed, dataPath[i], ctx); fails(e) {
			return
		}
		engines = append(engines, eng)
		if i == 0 {
			seed = transport.GetHostFirstMultiaddr(eng.Listener.Host)
		}
		go eng.Start()
	}
	cleanup = func() {
		for i := range engines {
			if engines[i] != nil {
				engines[i].Shutdown()
			}
		}
		for i := range dataPath {
			if dataPath[i] != "" {
				fails(os.RemoveAll(dataPath[i]))
			}
		}
	}
	return
}
