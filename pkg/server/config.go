package server

import (
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/multiformats/go-multiaddr"
)

func NewMultiAddr(addr string) (maddr multiaddr.Multiaddr) {

	var err error

	if maddr, err = multiaddr.NewMultiaddr(addr); check(err) {
		panic("Not a valid multiaddress.")
	}

	return
}

var DefaultConfig = &Config{
	ListenAddresses: []multiaddr.Multiaddr{},
	SeedAddresses: []multiaddr.Multiaddr{},
}

type Config struct {

	PrivKey crypto.PrivKey

	PublicAddress   multiaddr.Multiaddr
	SeedAddresses []multiaddr.Multiaddr
	ListenAddresses []multiaddr.Multiaddr
}
