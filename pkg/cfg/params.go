package cfg

import (
	"github.com/Indra-Labs/indra"
	log2 "github.com/Indra-Labs/indra/pkg/proc/pkg/log"
	"github.com/Indra-Labs/indra/pkg/wire"
	"github.com/multiformats/go-multiaddr"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

type Params struct {

	// Name defines a human-readable identifier for the network
	Name string

	// Net is a uint32 magic byte identifier for the network
	Net wire.IndraNet

	// DefaultPort is the default port for p2p listening
	DefaultPort string

	// SeedAddresses is a list of DNS hostnames used to bootstrap a new node on the network
	SeedAddresses   []string
}

func(self *Params) ParseSeedMultiAddresses() (addresses []multiaddr.Multiaddr, err error) {

	var adr multiaddr.Multiaddr

	for _, addr := range self.SeedAddresses {

		if adr, err = multiaddr.NewMultiaddr(addr+":"+self.DefaultPort); check(err) {
			return
		}

		addresses = append(addresses, adr)
	}

	return
}

var MainNetServerParams = &Params{

	Name: "mainnet",

	Net: wire.MainNet,

	DefaultPort: "8337",

	SeedAddresses: []string{
		// "seed0.indra.org",
		// "seed1.indra.org",
		// "seed2.indra.org",
		// "seed3.indra.org",
		// "seed0.example.com",
		// "seed1.example.com",
		// "seed2.example.com",
		// "seed3.example.com",
		// "1.1.1.1",
		// "::1",
	},

}

var TestNetServerParams = &Params{

	Name: "testnet",

	Net: wire.TestNet,

	DefaultPort: "58337",

	SeedAddresses: []string{
		// "seed0.testnet.indra.org",
		// "seed1.testnet.indra.org",
		// "seed2.testnet.indra.org",
		// "seed3.testnet.indra.org",
		// "seed0.testnet.example.com",
		// "seed1.testnet.example.com",
		// "seed2.testnet.example.com",
		// "seed3.testnet.example.com",
		// "1.1.1.1",
		// "::1",
	},

}

var SimnetServerParams = &Params{

	Name: "simnet",

	Net: wire.SimNet,

	DefaultPort: "62134",

	SeedAddresses:   []string{
		// We likely will never need any seed addresses here
	},
}
