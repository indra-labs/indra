package metrics

import (
	"context"
	"time"

	"github.com/indra-labs/indra"
	log2 "github.com/indra-labs/indra/pkg/log"
	"github.com/libp2p/go-libp2p/core/host"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

var (
	hostStatusTimeout = 30 * time.Second
)

func SetTimeout(key string, timeout time.Duration) {
	hostStatusTimeout = timeout
}

func HostStatus(ctx context.Context, host host.Host) {

	for {

		time.Sleep(hostStatusTimeout)

		select {

		case <-ctx.Done():

			log.I.Ln("shutting down metrics.hoststatus")

			return

		default:

		}

		log.I.Ln("---- host status ----")
		log.I.Ln("-- peers:", len(host.Network().Peers()))
		log.I.Ln("-- connections:", len(host.Network().Conns()))
		log.I.Ln("---- ---- ------ ----")
	}
}
