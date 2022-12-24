package server

import (
	"context"
	"github.com/Indra-Labs/indra"
	log2 "github.com/cybriq/proc/pkg/log"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/multiformats/go-multiaddr"
	dht "github.com/libp2p/go-libp2p-kad-dht"
)

var (
	log      = log2.GetLogger(indra.PathBase)
	check    = log.E.Chk
)

var DefaultServerConfig = Config {

	ListenAddresses: []multiaddr.Multiaddr{NewMultiAddrForced("/ip4/127.0.0.1/tcp/8337")},
	SeedAddresses: []multiaddr.Multiaddr{},
}

func NewMultiAddrForced(addr string) multiaddr.Multiaddr {

	var mta, _ = multiaddr.NewMultiaddr(addr)

	return mta
}

type Config struct {

	ListenAddresses []multiaddr.Multiaddr
	SeedAddresses []multiaddr.Multiaddr

}

type Server struct {

	host host.Host
	dht *dht.IpfsDHT
}

func (srv * Server) Serve() (err error) {
	return nil
}

func New(config Config) (srv *Server, err error) {

	// Start a new p2p host for the current node.
	log.D.Ln("generating a new p2p host.")

	log.I.Ln("p2p listeners:")
	for _, addr := range config.ListenAddresses {
		log.I.Ln("-", addr.String())
	}

	var p2pHost core.Host
	p2pHost, err = libp2p.New(libp2p.ListenAddrs(config.ListenAddresses...))

	if err != nil {
		return nil, err
	}

	log.I.Ln("host id:")
	log.I.Ln("-", p2pHost.ID())

	// Start a DHT, for use in peer discovery. We can't just make a new DHT
	// client because we want each peer to maintain its own local copy of the
	// DHT, so that the bootstrapping node of the DHT can go down without
	// inhibiting future peer discovery.
	ctx := context.Background()

	var kaDHT *dht.IpfsDHT

	if kaDHT, err = dht.New(ctx, p2pHost); err != nil {
		return nil, err
	}

	// Bootstrap the DHT. In the default configuration, this spawns a Background
	// thread that will refresh the peer table every five minutes.
	if err = kaDHT.Bootstrap(ctx); err != nil {
		return nil, err
	}

	s := Server {
		host: p2pHost,
		dht: kaDHT,
	}

	return &s, err
}
