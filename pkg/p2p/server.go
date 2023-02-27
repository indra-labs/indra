package p2p

import (
	"context"
	"git-indra.lan/indra-labs/indra/pkg/cfg"
	"git-indra.lan/indra-labs/indra/pkg/p2p/metrics"
	"github.com/libp2p/go-libp2p/core/crypto"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/multiformats/go-multiaddr"

	"git-indra.lan/indra-labs/indra"
	"git-indra.lan/indra-labs/indra/pkg/interrupt"
	"git-indra.lan/indra-labs/indra/pkg/p2p/introducer"
)

var (
	userAgent = "/indra:" + indra.SemVer + "/"
)

var (
	privKey         crypto.PrivKey
	p2pHost         host.Host
	seedAddresses   []multiaddr.Multiaddr
	listenAddresses []multiaddr.Multiaddr
	netParams       *cfg.Params
)

func init() {
	seedAddresses = []multiaddr.Multiaddr{}
	listenAddresses = []multiaddr.Multiaddr{}
}

type Server struct {
	context.Context

	config *Config

	host host.Host
}

func (srv *Server) Shutdown() (err error) {

	if err = srv.host.Close(); check(err) {
		// continue
	}

	log.I.Ln("shutdown complete")

	return
}

func (srv *Server) Serve() (err error) {

	log.I.Ln("starting the p2p server")

	// Here we create a context with cancel and add it to the interrupt handler
	var ctx context.Context
	var cancel context.CancelFunc

	ctx, cancel = context.WithCancel(context.Background())

	interrupt.AddHandler(cancel)

	// Introduce your node to the network
	go introducer.Bootstrap(ctx, srv.host, srv.config.SeedAddresses)

	// Get some basic metrics for the host
	// metrics.Init()
	// metrics.Set('indra.host.status.reporting.interval', 30 * time.Second)
	// metrics.Enable('indra.host.status')
	metrics.SetInterval(30 * time.Second)

	go metrics.HostStatus(ctx, srv.host)

	select {

	case <-ctx.Done():

		log.I.Ln("shutting down p2p server")

		srv.Shutdown()
	}

	return nil
}

func New(config *Config) (*Server, error) {

	log.I.Ln("initializing the p2p server")

	var err error
	var s Server

	s.config = config

	if s.host, err = libp2p.New(libp2p.Identity(config.PrivKey), libp2p.UserAgent(userAgent), libp2p.ListenAddrs(config.ListenAddresses...)); check(err) {
		return nil, err
	}

	log.I.Ln("host id:")
	log.I.Ln("-", s.host.ID())

	log.I.Ln("p2p listeners:")
	log.I.Ln("-", s.host.Addrs())

	if len(config.ConnectAddresses) > 0 {

		log.I.Ln("connect detected, using only the connect seed addresses")

		config.SeedAddresses = config.ConnectAddresses

		return &s, nil
	}

	var seedAddresses []multiaddr.Multiaddr

	if seedAddresses, err = config.Params.ParseSeedMultiAddresses(); check(err) {
		return nil, err
	}

	config.SeedAddresses = append(config.SeedAddresses, seedAddresses...)

	return &s, err
}
