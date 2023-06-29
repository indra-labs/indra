package introducer

import (
	"context"
	"errors"
	"fmt"
	record "github.com/libp2p/go-libp2p-record"
	"sync"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"

	log2 "github.com/indra-labs/indra/pkg/proc/log"
)

var (
	log   = log2.GetLogger()
	check = log.E.Chk
)

var (
	name                       = "[introducer.bootstrap]"
	protocolPrefix protocol.ID = "/indra"
)

var (
	wg sync.WaitGroup
	m  sync.Mutex
	c  context.Context
	h  host.Host = nil

	kadht          *dht.IpfsDHT
	bootstrapPeers []peer.AddrInfo
)

func Bootstrap(ctx context.Context, host host.Host,
	seeds []multiaddr.Multiaddr) (err error) {

	log.I.Ln("starting [introducer.bootstrap]")

	// Guarding against multiple instantiations
	if !m.TryLock() {
		return errors.New("[introducer.bootstrap] service is in use")
	}

	c = ctx
	h = host

	o := "using seeds:\n\n"

	var bootstrapPeer *peer.AddrInfo

	for _, seed := range seeds {

		o += fmt.Sprintln("-", seed.String())

		if bootstrapPeer, err = peer.AddrInfoFromP2pAddr(seed); check(err) {
			return
		}

		// We can skip ourselves
		if bootstrapPeer.ID == host.ID() {
			continue
		}

		bootstrapPeers = append(bootstrapPeers, *bootstrapPeer)
	}

	log.I.Ln(o)

	var options = []dht.Option{
		dht.Mode(dht.ModeServer),
		dht.ProtocolPrefix(protocolPrefix),
		dht.BootstrapPeers(bootstrapPeers...),
		dht.DisableValues(),
		dht.DisableProviders(),
		dht.Validator(record.PublicKeyValidator{}),
	}

	if kadht, err = dht.New(ctx, h, options...); check(err) {
		return
	}

	if err = kadht.Bootstrap(ctx); check(err) {
		return
	}

	log.I.Ln("[introducer.bootstrap] is ready")

	go func() {
		select {
		case <-c.Done():

			log.I.Ln("shutting down [introducer.bootstrap]")

			return
		}
	}()

	return
}
