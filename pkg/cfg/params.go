package cfg

import (
	"git.indra-labs.org/dev/ind/pkg/node"
	log2 "git.indra-labs.org/dev/ind/pkg/proc/log"
	"github.com/multiformats/go-multiaddr"
	"go.uber.org/atomic"
	"strconv"
)

var (
	log   = log2.GetLogger()
	check = log.E.Chk

	current atomic.String
)

func init() {
	// Set default so all code has *something* to use (mainly for tests).
	current.Store(TestNet)
}

// CurrentNetworkParams is set by the controlling server, but we set it to
// default at initialisation. There is no reason for this to not be a
// singleton variable. There should only be writes to this variable before
// startup.
func CurrentNetworkParams() *Params {
	return SelectNetworkParams(current.Load())
}

const (
	// MainNet is the identifier string for the main production network.
	MainNet = "mainnet"
	// TestNet is the identifier string for the test network.
	TestNet = "testnet"
	// SimNet is the identifier string for simulation networks.
	SimNet = "simnet"
)

// Params is the specification for an indranet swarm (mainnet, testnet, etc).
type Params struct {

	// Name defines a human-readable identifier for the network
	Name string

	// Net is a uint32 magic byte identifier for the network
	Net node.Swarm

	// DefaultPort is the default port for p2p listening
	DefaultPort string

	// DNSSeedAddresses is a list of DNS hostnames used to bootstrap a new node
	// on the network. todo: maybe DNS should not be assumed, this puts ICANN in control...
	DNSSeedAddresses []*SeedAddress
}

// SelectNetworkParams returns the network parameters associated with the
// network name.
//
// This should not be set except at startup.
func SelectNetworkParams(network string) *Params {

	if nw, ok := params[network]; ok {

		return nw
	}
	panic("invalid network, exiting...")

	return nil
}

func SetNetworkParams(network string) {

	if _, ok := params[network]; ok {
		current.Store(network)
	}
	panic("invalid network, exiting...")
}

func GetCurrentNetworkParams() *Params {
	return SelectNetworkParams(current.Load())
}

func GetCurrentDefaultPort() uint16 {
	v, _ := strconv.Atoi(SelectNetworkParams(current.Load()).
		DefaultPort)
	return uint16(v)
}

// ParseSeedMultiAddresses returns the addresses of the seeds as a slice of multiaddr.Multiaddr.
func (p *Params) ParseSeedMultiAddresses() (addresses []multiaddr.Multiaddr, err error) {

	var adr multiaddr.Multiaddr

	addresses = []multiaddr.Multiaddr{}

	for _, addr := range p.DNSSeedAddresses {

		if adr, err = multiaddr.NewMultiaddr("/dns4/" +
			addr.DNSAddress + "/tcp/" + p.DefaultPort + "/p2p/" +
			addr.ID); check(err) {

			return
		}

		addresses = append(addresses, adr)
	}

	return
}

// GetSeedsMultiAddrStrings returns the seeds multiaddrs as encoded strings.
func (p *Params) GetSeedsMultiAddrStrings() (seeds []string) {
	for _, addr := range p.DNSSeedAddresses {
		seeds = append(seeds, "/dns4/"+addr.DNSAddress+"/tcp/"+
			p.DefaultPort+"/p2p/"+addr.ID)
	}
	return
}

var (
	params = map[string]*Params{
		MainNet: {

			Name: "mainnet",

			Net: node.MainNet,

			DefaultPort: "8337",

			DNSSeedAddresses: []*SeedAddress{
				NewSeedAddress("seed0.indra.org", "12D3KooWCfTmWavthiVV7Vkm9eouCdiLdGnhd2PShQ2hiu2VVU6Q"),
				NewSeedAddress("seed1.indra.org", "12D3KooWASwYWP2gMh581EQG25nauvWfwAU3g6v8TugEoEzL5Ags"),
				NewSeedAddress("seed2.indra.org", "12D3KooWFW7k2YcxjZrqWXJhmoCTNiNtgjLkEUeqgvZRAF3xHZjs"),
				NewSeedAddress("seed3.indra.org", "12D3KooWPxx3WMiCv3SwBNfrM6peGBWDypJqqxfdGgZKpr7BF9Vo"),
				// NewSeedAddress("seed0.example.com", "12D3KooWDj2wXRVPRVP8HcQXTyAXeigAAjaX6hgdgALyNFuK1Htv"),
				// NewSeedAddress("seed1.example.com", "12D3KooWMkBp6E2qjz2saq9eocT9FTh3zuoP5yAcFgFGSfXoZN8K"),
				// NewSeedAddress("seed2.example.com", "12D3KooWEonhWcCp6FMwycNFrE5hSDbPdezy5ftBcHLxLPoESzgZ"),
				// NewSeedAddress("seed3.example.com", "12D3KooWFq8irCNNCdE4zxjcUGVdG47fnPSd4hj9MsxH8RAunHTx"),
			},
		},
		TestNet: {

			Name: "testnet",

			Net: node.TestNet,

			DefaultPort: "58337",

			DNSSeedAddresses: []*SeedAddress{
				// NewSeedAddress("seed0.indra.org", "12D3KooWCfTmWavthiVV7Vkm9eouCdiLdGnhd2PShQ2hiu2VVU6Q"),
				// NewSeedAddress("seed1.indra.org", "12D3KooWASwYWP2gMh581EQG25nauvWfwAU3g6v8TugEoEzL5Ags"),
				// NewSeedAddress("seed2.indra.org", "12D3KooWFW7k2YcxjZrqWXJhmoCTNiNtgjLkEUeqgvZRAF3xHZjs"),
				// NewSeedAddress("seed3.indra.org", "12D3KooWPxx3WMiCv3SwBNfrM6peGBWDypJqqxfdGgZKpr7BF9Vo"),
				// NewSeedAddress("seed0.example.com", "12D3KooWDj2wXRVPRVP8HcQXTyAXeigAAjaX6hgdgALyNFuK1Htv"),
				// NewSeedAddress("seed1.example.com", "12D3KooWMkBp6E2qjz2saq9eocT9FTh3zuoP5yAcFgFGSfXoZN8K"),
				// NewSeedAddress("seed2.example.com", "12D3KooWEonhWcCp6FMwycNFrE5hSDbPdezy5ftBcHLxLPoESzgZ"),
				// NewSeedAddress("seed3.example.com", "12D3KooWFq8irCNNCdE4zxjcUGVdG47fnPSd4hj9MsxH8RAunHTx"),
			},
		},
		SimNet: {

			Name: "simnet",

			Net: node.SimNet,

			DefaultPort: "62134",

			// Should be passed via --seed
			DNSSeedAddresses: []*SeedAddress{
				NewSeedAddress("seed0", "16Uiu2HAmHL9cwDdGdGEQk7K5xTBH3rUKoRWS3tWgjej9WmjMG6L9"),
				NewSeedAddress("seed1", "16Uiu2HAm3ZVpmNPnk67eCKsR3HCMSLhbw3THcBDpyKwCE5GNgHVN"),
				NewSeedAddress("seed2", "16Uiu2HAmTP7o9yyuFz8f5sRGr7KTamuMo6UguuTqPMaJavdGX7Ju"),
			},
		},
	}
)
