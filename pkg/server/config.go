package server

import "github.com/multiformats/go-multiaddr"

var DefaultServerConfig = Config{

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
