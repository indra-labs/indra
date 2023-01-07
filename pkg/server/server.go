package server

import (
	"context"
	"github.com/indra-labs/indra"
	"github.com/indra-labs/indra/pkg/cfg"
	"github.com/indra-labs/indra/pkg/interrupt"
	log2 "github.com/indra-labs/indra/pkg/log"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
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
	userAgent = "/indra:"+indra.SemVer+"/"
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

	log.I.Ln("shutting down the dht...")

	if srv.dht.Close(); check(err) {
		return
	}

	log.I.Ln("shutting down the p2p host...")

	if srv.host.Close(); check(err) {
		return
	}

	log.I.Ln("shutdown complete.")

	return nil
}

func seedConnect(ctx context.Context, attempts int) {

}

func peer_metrics(host host.Host, quitChan <-chan struct{}) {

	for {

		select {
		case <- quitChan:
			break
		default:

		}

		log.I.Ln("peers:",len(host.Network().Peers()))

		time.Sleep(10 * time.Second)
	}
}

func (srv *Server) attempt(ctx context.Context, peer *peer.AddrInfo, attempts_left int, wg sync.WaitGroup) {

	log.I.Ln("attempting connection to", peer.ID)

	defer wg.Done()

	if err := srv.host.Connect(srv.Context, *peer); check(err) {

		log.E.Ln("connection attempt failed to", peer.ID)

		attempts_left--

		if attempts_left <= 0 {

			return
		}

		time.Sleep(10 * time.Second)

		srv.attempt(ctx, peer, attempts_left, wg)

		return
	}

	log.I.Ln("connection established with seed node:", peer.ID)

	ctx.Done()
}

func (srv *Server) seed_connect() (err error) {

	log.I.Ln("attempting to peer with seed addresses...")

	// We will first attempt to seed_connect to the seed addresses.
	var wg sync.WaitGroup

	var peerInfo *peer.AddrInfo

	for _, peerAddr := range srv.config.SeedAddresses {

		log.I.Ln("-", peerAddr.String())

		if peerInfo, err = peer.AddrInfoFromP2pAddr(peerAddr); check(err) {
			return
		}

		if peerInfo.ID == srv.host.ID() {

			log.I.Ln("attempting to seed_connect to self, skipping...")

			continue
		}

		wg.Add(1)

		ctx, cancel := context.WithTimeout(context.Background(), 30 * time.Second)
		defer cancel()

		go srv.attempt(ctx, peerInfo, 3, wg)
	}

	wg.Wait()

	return
}

func (srv *Server) Serve() (err error) {

	go peer_metrics(srv.host, srv.Context.Done())

	//log.I.Ln("bootstrapping the DHT")

	// Bootstrap the DHT. In the default configuration, this spawns a Background
	// thread that will refresh the peer table every five minutes.
	if err = srv.dht.Bootstrap(srv.Context); check(err) {
		return err
	}

	if err = srv.seed_connect(); check(err) {
		return
	}

	select {
	case <-srv.Context.Done():
		srv.Shutdown()
	}

	return nil
}

func New(params *cfg.Params, config *Config) (srv *Server, err error) {

	log.I.Ln("initializing the server.")

	var s Server

	s.params = params
	s.config = config

	var cancel context.CancelFunc

	s.Context, cancel = context.WithCancel(context.Background())

	// Add an interrupt handler for the server shutdown
	interrupt.AddHandler(cancel)

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

	log.I.Ln("seed addresses:")
	log.I.Ln("-", config.SeedAddresses)

	// Start a DHT, for use in peer discovery. We can't just make a new DHT
	// client because we want each peer to maintain its own local copy of the
	// DHT, so that the bootstrapping node of the DHT can go down without
	// inhibiting future peer discovery.
	//if s.dht, err = dht.New(s.Context, s.host); check(err) {
	//	return nil, err
	//}

	return &s, err
}
