package server

import (
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
}

type Config struct {

	PublicAddress   multiaddr.Multiaddr
	ListenAddresses []multiaddr.Multiaddr
}
