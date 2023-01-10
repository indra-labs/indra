package server

import (
	"context"
	"github.com/indra-labs/indra"
	"github.com/indra-labs/indra/pkg/cfg"
	"github.com/indra-labs/indra/pkg/interrupt"
	log2 "github.com/indra-labs/indra/pkg/log"
	"github.com/indra-labs/indra/pkg/p2p/introducer"
	"github.com/indra-labs/indra/pkg/p2p/metrics"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/multiformats/go-multiaddr"
	"time"
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
}

func (srv *Server) Restart() (err error) {

	log.I.Ln("restarting the server.")

	return nil
}

func (srv *Server) Shutdown() (err error) {

	log.I.Ln("shutting down [p2p.host]")

	if srv.host.Close(); check(err) {
		return
	}

	log.I.Ln("shutdown complete")

	return nil
}

func (srv *Server) Serve() (err error) {

	log.I.Ln("starting the server")

	// Here we create a context with cancel and add it to the interrupt handler
	var ctx context.Context
	var cancel context.CancelFunc

	ctx, cancel = context.WithCancel(context.Background())

	interrupt.AddHandler(cancel)

	// Introduce your node to the network
	go introducer.Bootstrap(ctx, srv.host, srv.config.SeedAddresses)

	// Get some basic metrics for the host
	//metrics.Init()
	//metrics.Set('indra.host.status.reporting.interval', 30 * time.Second)
	//metrics.Enable('indra.host.status')
	metrics.SetInterval(30 * time.Second)

	go metrics.HostStatus(ctx, srv.host)

	select {

	case <-ctx.Done():

		log.I.Ln("shutting down server")

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

	return &s, err
}
