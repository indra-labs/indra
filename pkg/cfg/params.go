package cfg

import (
	"github.com/indra-labs/indra"
	log2 "github.com/indra-labs/indra/pkg/log"
	"github.com/indra-labs/indra/pkg/wire"
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

	// DNSSeedAddresses is a list of DNS hostnames used to bootstrap a new node on the network
	DNSSeedAddresses []*DNSSeedAddress
}

func (self *Params) ParseSeedMultiAddresses() (addresses []multiaddr.Multiaddr, err error) {

	var adr multiaddr.Multiaddr

	addresses = []multiaddr.Multiaddr{}

	for _, addr := range self.DNSSeedAddresses {

		if adr, err = multiaddr.NewMultiaddr("/dns4/" + addr.DNSAddress + "/tcp/" + self.DefaultPort + "/p2p/" + addr.ID); check(err) {
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

	DNSSeedAddresses: []*DNSSeedAddress{
		NewSeedAddress("seed0.indra.org", "12D3KooWCfTmWavthiVV7Vkm9eouCdiLdGnhd2PShQ2hiu2VVU6Q"),
		NewSeedAddress("seed1.indra.org", "12D3KooWASwYWP2gMh581EQG25nauvWfwAU3g6v8TugEoEzL5Ags"),
		NewSeedAddress("seed2.indra.org", "12D3KooWFW7k2YcxjZrqWXJhmoCTNiNtgjLkEUeqgvZRAF3xHZjs"),
		NewSeedAddress("seed3.indra.org", "12D3KooWPxx3WMiCv3SwBNfrM6peGBWDypJqqxfdGgZKpr7BF9Vo"),
		// NewSeedAddress("seed0.example.com", "12D3KooWDj2wXRVPRVP8HcQXTyAXeigAAjaX6hgdgALyNFuK1Htv"),
		// NewSeedAddress("seed1.example.com", "12D3KooWMkBp6E2qjz2saq9eocT9FTh3zuoP5yAcFgFGSfXoZN8K"),
		// NewSeedAddress("seed2.example.com", "12D3KooWEonhWcCp6FMwycNFrE5hSDbPdezy5ftBcHLxLPoESzgZ"),
		// NewSeedAddress("seed3.example.com", "12D3KooWFq8irCNNCdE4zxjcUGVdG47fnPSd4hj9MsxH8RAunHTx"),
	},
}

var TestNetServerParams = &Params{

	Name: "testnet",

	Net: wire.TestNet,

	DefaultPort: "58337",

	DNSSeedAddresses: []*DNSSeedAddress{
		// NewSeedAddress("seed0.indra.org", "12D3KooWCfTmWavthiVV7Vkm9eouCdiLdGnhd2PShQ2hiu2VVU6Q"),
		// NewSeedAddress("seed1.indra.org", "12D3KooWASwYWP2gMh581EQG25nauvWfwAU3g6v8TugEoEzL5Ags"),
		// NewSeedAddress("seed2.indra.org", "12D3KooWFW7k2YcxjZrqWXJhmoCTNiNtgjLkEUeqgvZRAF3xHZjs"),
		// NewSeedAddress("seed3.indra.org", "12D3KooWPxx3WMiCv3SwBNfrM6peGBWDypJqqxfdGgZKpr7BF9Vo"),
		// NewSeedAddress("seed0.example.com", "12D3KooWDj2wXRVPRVP8HcQXTyAXeigAAjaX6hgdgALyNFuK1Htv"),
		// NewSeedAddress("seed1.example.com", "12D3KooWMkBp6E2qjz2saq9eocT9FTh3zuoP5yAcFgFGSfXoZN8K"),
		// NewSeedAddress("seed2.example.com", "12D3KooWEonhWcCp6FMwycNFrE5hSDbPdezy5ftBcHLxLPoESzgZ"),
		// NewSeedAddress("seed3.example.com", "12D3KooWFq8irCNNCdE4zxjcUGVdG47fnPSd4hj9MsxH8RAunHTx"),
	},
}

var SimnetServerParams = &Params{

	Name: "simnet",

	Net: wire.SimNet,

	DefaultPort: "62134",

	// Should be passed via --seed
	DNSSeedAddresses: []*DNSSeedAddress{
		NewSeedAddress("seed0", "16Uiu2HAmCxWoKp4vs7xrmzbScHEhUK7trCgCPhKPZRBiUvSxS7xA"),
		NewSeedAddress("seed1", "16Uiu2HAmTKk6BvJFPmcQ6q92XgvQ4ZPu1AVjQxMvCfM4you9Zyvc"),
		NewSeedAddress("seed2", "16Uiu2HAm8tCAW7D9WFLxkda52R73nSk9yBCFW8uwA4MZPzHYVhnW"),
	},
}
