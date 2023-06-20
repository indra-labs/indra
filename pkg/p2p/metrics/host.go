package metrics

import (
	"context"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/host"

	log2 "github.com/indra-labs/indra/pkg/proc/log"
)

var (
	log   = log2.GetLogger()
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

				log.I.Ln(`
---- host status ----
-- peers:`, len(host.Network().Peers()),`
-- connections:`, len(host.Network().Conns()),`
`)

			case <-ctx.Done():

				log.I.Ln("shutting down [metrics.hoststatus]")

				return
			}
		}
	}()
}
