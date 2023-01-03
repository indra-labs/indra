package server

import "github.com/multiformats/go-multiaddr"

var (
	seedList = []string{
		//"/dns4/seed0.indra.org/tcp/8337",
		//"/dns4/seed1.indra.org/tcp/8337",
		//"/dns4/seed2.indra.org/tcp/8337",
		//"/dns4/seed3.indra.org/tcp/8337",
		//"/dns4/seed0.example.com/tcp/8337",
		//"/dns4/seed1.example.com/tcp/8337",
		//"/dns4/seed2.example.com/tcp/8337",
		//"/dns4/seed3.example.com/tcp/8337",
		//"/ip4/1.1.1.1/tcp/8337",
		//"/ip6/::1/tcp/3217",
	}
)

var DefaultServerConfig = Config{

	SeedAddresses:   []multiaddr.Multiaddr{},
	ListenAddresses: []multiaddr.Multiaddr{},
}

var SimnetServerConfig = Config{

	SeedAddresses:   []multiaddr.Multiaddr{},
	ListenAddresses: []multiaddr.Multiaddr{},
}

func NewMultiAddrForced(addr string) multiaddr.Multiaddr {

	var mta, _ = multiaddr.NewMultiaddr(addr)

	return mta
}

type Config struct {
	PublicAddress   multiaddr.Multiaddr
	ListenAddresses []multiaddr.Multiaddr
	SeedAddresses   []multiaddr.Multiaddr
}
