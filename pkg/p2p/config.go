package p2p

import (
	"git.indra-labs.org/dev/ind/pkg/cfg"
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
	ListenAddresses:  []multiaddr.Multiaddr{},
	SeedAddresses:    []multiaddr.Multiaddr{},
	ConnectAddresses: []multiaddr.Multiaddr{},
}

type Config struct {
	PrivKey crypto.PrivKey

	PublicAddress    multiaddr.Multiaddr
	SeedAddresses    []multiaddr.Multiaddr
	ConnectAddresses []multiaddr.Multiaddr
	ListenAddresses  []multiaddr.Multiaddr

	Params *cfg.Params
}

func (c *Config) SetNetwork(network string) {

	c.Params = cfg.SelectNetworkParams(network)
}
