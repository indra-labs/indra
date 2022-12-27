package server

import "github.com/multiformats/go-multiaddr"

var DefaultServerConfig = Config{

	ListenAddresses: []multiaddr.Multiaddr{NewMultiAddrForced("/ip4/127.0.0.1/tcp/8337")},
	SeedAddresses:   []multiaddr.Multiaddr{},
}

func NewMultiAddrForced(addr string) multiaddr.Multiaddr {

	var mta, _ = multiaddr.NewMultiaddr(addr)

	return mta
}

type Config struct {

	PublicAddress multiaddr.Multiaddr
	ListenAddresses []multiaddr.Multiaddr
	SeedAddresses   []multiaddr.Multiaddr



}
