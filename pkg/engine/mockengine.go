package engine

import (
	"context"
	"crypto/rand"
	badger "github.com/indra-labs/go-ds-badger3"
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/sha256"
	"github.com/indra-labs/indra/pkg/engine/node"
	"github.com/indra-labs/indra/pkg/engine/transport"
	"github.com/indra-labs/indra/pkg/util/multi"
	"github.com/multiformats/go-multiaddr"
	"net/netip"
	"os"
)

// CreateMockEngine creates an indra Engine with a random localhost listener.
func CreateMockEngine(seed []string, dataPath string, ctx context.Context, cancel context.CancelFunc) (ng *Engine, store *badger.Datastore, closer func()) {

	var (
		e    error
		keys []*crypto.Keys
		k    *crypto.Keys
	)
	defer func(f *error) {
		if *f != nil {
			fails(os.RemoveAll(dataPath))
		}
	}(&e)
	if k, e = crypto.GenerateKeys(); fails(e) {
		return
	}
	keys = append(keys, k)

	secret := sha256.New()
	rand.Read(secret[:])
	store, closer = transport.BadgerStore(dataPath, secret[:])
	if store == nil {
		log.E.Ln("could not open database")
		return nil, store, closer
	}

	var l *transport.Listener
	if l, e = transport.NewListener(seed,
		[]string{transport.LocalhostZeroIPv4TCP,
			transport.LocalhostZeroIPv6TCP}, k, store, closer, ctx,
		transport.DefaultMTU, cancel); fails(e) {
		return
	}
	if l == nil {
		panic("maybe you have no network device?")
	}

	sa := transport.GetHostFirstMultiaddr(l.Host)

	var ma multiaddr.Multiaddr
	if ma, e = multiaddr.NewMultiaddr(sa); fails(e) {
		return
	}

	var ap netip.AddrPort
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

func CreateAndStartMockEngines(n int, ctx context.Context,
	cancel context.CancelFunc) (engines []*Engine, closer func(), e error) {

	closer = func() {}
	var seed []string
	dataPath := make([]string, n)
	for i := 0; i < n; i++ {
		dataPath[i], e = os.MkdirTemp(os.TempDir(), "badger")
		if e != nil {
			return
		}
		var eng *Engine
		if eng, _, _ = CreateMockEngine(seed, dataPath[i], ctx, cancel); fails(e) {
			return
		}
		engines = append(engines, eng)
		if i == 0 {
			seed = transport.GetHostMultiaddrs(eng.Listener.Host)
		}
		go eng.Start()
	}
	closer = func() {
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
