package seed

import (
	"context"
	"errors"
	"github.com/indra-labs/indra"
	log2 "github.com/indra-labs/indra/pkg/log"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"sync"
	"time"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

var (
	defaultConnectionAttempts        uint = 3
	defaultConnectionAttemptInterval      = 3 * time.Second

	defaultConnectionsMax       uint = 32
	defaultConnectionsToSatisfy uint = 5
)

var (
	wg sync.WaitGroup
	m  sync.Mutex
	c  context.Context
	h  host.Host = nil

	failedChan = make(chan error)
)

func connection_attempt(peer *peer.AddrInfo, attempts_left uint) {

	if attempts_left == 0 {

		wg.Done()

		return
	}

	log.I.Ln("attempting connection", peer.ID)

	if err := h.Connect(c, *peer); err != nil {

		log.I.Ln("connection attempt failed:", peer.ID)

		select {

		case <-time.After(defaultConnectionAttemptInterval):

			connection_attempt(peer, attempts_left-1)

		case <-c.Done():

			log.I.Ln("[seed.bootstrap] connection to", peer.ID ,"interrupted, shutting down")

			wg.Done()
		}

		return
	}

	log.I.Ln("seed connection established:", peer.String())

	wg.Done()
}

func Bootstrap(ctx context.Context, host host.Host, seeds []multiaddr.Multiaddr) (err error) {

	log.I.Ln("starting [seed.bootstrap]")

	// Guarding against multiple instantiations
	if !m.TryLock() {
		return errors.New("bootstrapping service is in use.")
	}

	c = ctx
	h = host

	log.I.Ln("attempting peering with seeds...")

	var peerInfo *peer.AddrInfo

	for _, peerAddr := range seeds {

		log.I.Ln("-", peerAddr.String())

		if peerInfo, err = peer.AddrInfoFromP2pAddr(peerAddr); check(err) {
			return
		}

		// We can skip ourselves
		if peerInfo.ID == host.ID() {
			continue
		}

		wg.Add(1)

		go connection_attempt(peerInfo, defaultConnectionAttempts)
	}

	wg.Wait()

	select {
	case <- c.Done():

		log.I.Ln("shutting down [seed.bootstrap]")

		return
	}

	log.I.Ln("finished seed bootstrapping")

	return
}
