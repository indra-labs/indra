package server

import (
	"context"
	"github.com/Indra-Labs/indra"
	"github.com/Indra-Labs/indra/pkg/cfg"
	"github.com/cybriq/proc/pkg/interrupt"
	log2 "github.com/cybriq/proc/pkg/log"
	"github.com/davecgh/go-spew/spew"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"sync"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

type Server struct {
	context.Context

	config Config

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

func (srv *Server) Serve() (err error) {

	// Bootstrap the DHT. In the default configuration, this spawns a Background
	// thread that will refresh the peer table every five minutes.
	if err = srv.dht.Bootstrap(srv.Context); check(err) {
		return err
	}

	select {
	case <-srv.Context.Done():
		srv.Shutdown()
	}

	// We will first attempt to connect to the seed addresses.
	var wg sync.WaitGroup

	spew.Dump(srv.params.ParseSeedMultiAddresses())

	var seedAddresses []multiaddr.Multiaddr

	seedAddresses, err = srv.params.ParseSeedMultiAddresses()

	for _, peerAddr := range seedAddresses {
		peerinfo, _ := peer.AddrInfoFromP2pAddr(peerAddr)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := srv.host.Connect(srv, *peerinfo); err != nil {
				log.W.Ln(err)
			} else {
				log.I.Ln("Connection established with bootstrap node:", *peerinfo)
			}
		}()
	}

	wg.Wait()

	return nil
}

func New(params *cfg.Params, config *Config) (srv *Server, err error) {

	log.I.Ln("initializing the server.")

	var s Server

	s.params = params

	var cancel context.CancelFunc

	s.Context, cancel = context.WithCancel(context.Background())

	// Add an interrupt handler for the server shutdown
	interrupt.AddHandler(cancel)

	if s.host, err = libp2p.New(libp2p.ListenAddrs(config.ListenAddresses...)); check(err) {
		return nil, err
	}

	log.I.Ln("p2p listeners:")
	log.I.Ln("-", s.host.Addrs())

	log.I.Ln("host id:")
	log.I.Ln("-", s.host.ID())

	// Start a DHT, for use in peer discovery. We can't just make a new DHT
	// client because we want each peer to maintain its own local copy of the
	// DHT, so that the bootstrapping node of the DHT can go down without
	// inhibiting future peer discovery.
	if s.dht, err = dht.New(s.Context, s.host); check(err) {
		return nil, err
	}

	return &s, err
}
