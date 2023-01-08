package server

import (
	"context"
	"github.com/indra-labs/indra"
	"github.com/indra-labs/indra/pkg/cfg"
	"github.com/indra-labs/indra/pkg/interrupt"
	log2 "github.com/indra-labs/indra/pkg/log"
	"github.com/indra-labs/indra/pkg/p2p/metrics"
	"github.com/indra-labs/indra/pkg/p2p/seed"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/multiformats/go-multiaddr"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

var (
	userAgent = "/indra:" + indra.SemVer + "/"
)

type Server struct {
	context.Context

	config *Config

	params *cfg.Params

	host host.Host
	dht  *dht.IpfsDHT
}

func (srv *Server) Restart() (err error) {

	log.I.Ln("restarting the server.")

	return nil
}

func (srv *Server) Shutdown() (err error) {

	//log.I.Ln("shutting down the dht...")
	//
	//if srv.dht.Close(); check(err) {
	//	return
	//}

	log.I.Ln("shutting down [p2p.host]")

	if srv.host.Close(); check(err) {
		return
	}

	log.I.Ln("shutdown complete")

	return nil
}

func (srv *Server) Serve() (err error) {

	// Here we create a context with cancel and add it to the interrupt handler
	var ctx context.Context
	var cancel context.CancelFunc

	ctx, cancel = context.WithCancel(context.Background())

	interrupt.AddHandler(cancel)

	// Get some basic metrics for the host
	//metrics.Init()
	//metrics.Set('indra.host.status.reporting.interval', 30 * time.Second)
	//metrics.Enable('indra.host.status')
	go metrics.HostStatus(ctx, srv.host)

	// Run the bootstrapping service on the peer.
	go seed.Bootstrap(ctx, srv.host, srv.config.SeedAddresses)

	//log.I.Ln("bootstrapping the DHT")

	// Bootstrap the DHT. In the default configuration, this spawns a Background
	// thread that will refresh the peer table every five minutes.
	//if err = srv.dht.Bootstrap(srv.Context); check(err) {
	//	return err
	//}

	//log.I.Ln("successfully connected")

	//var pingService *ping.PingService
	//
	//if pingService = ping.NewPingService(srv.host); check(err) {
	//	return
	//}
	//
	//go func() {
	//
	//	log.I.Ln("attempting ping")
	//
	//	for {
	//
	//		for _, peer := range srv.host.Peerstore().Peers() {
	//
	//			select {
	//				case result := <- pingService.Ping(context.Background(), peer):
	//					log.I.Ln("ping", peer.String(), "-", result.RTT)
	//			}
	//		}
	//
	//		time.Sleep(10 * time.Second)
	//	}
	//
	//}()

	select {
	case <-ctx.Done():
		srv.Shutdown()
	}

	return nil
}

func New(params *cfg.Params, config *Config) (srv *Server, err error) {

	log.I.Ln("initializing the server")

	var s Server

	s.params = params
	s.config = config

	if s.host, err = libp2p.New(libp2p.Identity(config.PrivKey), libp2p.UserAgent(userAgent), libp2p.ListenAddrs(config.ListenAddresses...)); check(err) {
		return nil, err
	}

	log.I.Ln("host id:")
	log.I.Ln("-", s.host.ID())

	log.I.Ln("p2p listeners:")
	log.I.Ln("-", s.host.Addrs())

	var seedAddresses []multiaddr.Multiaddr

	if seedAddresses, err = params.ParseSeedMultiAddresses(); check(err) {
		return
	}

	config.SeedAddresses = append(config.SeedAddresses, seedAddresses...)

	// Start a DHT, for use in peer discovery. We can't just make a new DHT
	// client because we want each peer to maintain its own local copy of the
	// DHT, so that the bootstrapping node of the DHT can go down without
	// inhibiting future peer discovery.
	//if s.dht, err = dht.New(s.Context, s.host); check(err) {
	//	return nil, err
	//}

	return &s, err
}
