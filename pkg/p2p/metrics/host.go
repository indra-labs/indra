package metrics

import (
	"context"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/host"

	"git-indra.lan/indra-labs/indra"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

var (
	hostStatusInterval = 10 * time.Second
)

var (
	mutex sync.Mutex
)

func SetInterval(timeout time.Duration) {
	hostStatusInterval = timeout
}

func HostStatus(ctx context.Context, host host.Host) {

	log.I.Ln("starting [metrics.hoststatus]")

	// Guarding against multiple instantiations
	if !mutex.TryLock() {
		return
	}

	log.I.Ln("[metrics.hoststatus] is ready")

	go func() {
		for {
			select {
			case <-time.After(hostStatusInterval):

				log.I.Ln()
				log.I.Ln("---- host status ----")
				log.I.Ln("-- peers:", len(host.Network().Peers()))
				log.I.Ln("-- connections:", len(host.Network().Conns()))
				log.I.Ln("---- ---- ------ ----")

			case <-ctx.Done():

				log.I.Ln("shutting down [metrics.hoststatus]")

				return
			}
		}
	}()
}
