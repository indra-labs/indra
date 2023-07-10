package engine

import (
	"context"
	"errors"
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/engine/node"
	"github.com/indra-labs/indra/pkg/engine/transport"
	"github.com/indra-labs/indra/pkg/util/multi"
	"github.com/multiformats/go-multiaddr"
	"net/netip"
	"os"
)

// CreateMockEngine creates an indra Engine with a random localhost listener.
func CreateMockEngine(seed, dataPath string) (ng *Engine, cancel func(), e error) {
	defer func(f *error) {
		if *f != nil {
			fails(os.RemoveAll(dataPath))
		}
	}(&e)
	var ctx context.Context
	ctx, cancel = context.WithCancel(context.Background())
	var keys []*crypto.Keys
	var k *crypto.Keys
	if k, e = crypto.GenerateKeys(); fails(e) {
		return
	}
	keys = append(keys, k)
	var l *transport.Listener
	if l, e = transport.NewListener([]string{seed},
		[]string{transport.LocalhostZeroIPv4TCP, transport.LocalhostZeroIPv6TCP}, dataPath, k, ctx,
		transport.DefaultMTU); fails(e) {
		return
	}
	if l == nil {
		cancel()
		return nil, nil, errors.New("got nil listener")
	}
	sa := transport.GetHostAddress(l.Host)
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

func CreateAndStartMockEngines(n int) (engines []*Engine, cleanup func(), e error) {
	cleanup = func() {}
	var seed string
	for i := 0; i < n; i++ {
		dataPath, err := os.MkdirTemp(os.TempDir(), "badger")
		if err != nil {
			cleanup()
			return
		}
		var eng *Engine
		if eng, _, e = CreateMockEngine(seed, dataPath); fails(e) {
			cleanup()
			return
		}
		engines = append(engines, eng)
		if i == 0 {
			seed = transport.GetHostAddress(eng.Listener.Host)
		}
		cleanup = func() {
			cleanup()
			fails(os.RemoveAll(dataPath))
		}
		go eng.Start()
	}
	return
}
