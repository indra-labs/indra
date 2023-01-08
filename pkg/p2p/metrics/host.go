package metrics

import (
	"context"
	"github.com/indra-labs/indra"
	log2 "github.com/indra-labs/indra/pkg/log"
	"github.com/libp2p/go-libp2p/core/host"
	"time"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

var (
	hostStatusInterval = 10 * time.Second
)

func SetTimeout(key string, timeout time.Duration) {
	hostStatusInterval = timeout
}

func HostStatus(ctx context.Context, host host.Host) {

	log.I.Ln("starting [metrics.hoststatus]")

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
}
