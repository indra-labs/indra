package seed

import (
	"git-indra.lan/indra-labs/indra/pkg/cfg"
	"git-indra.lan/indra-labs/indra/pkg/rpc"
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
	RPCConfig: &rpc.RPCConfig{
		Key:            &rpc.DefaultRPCPrivateKey,
		ListenPort:     0,
		Peer_Whitelist: []rpc.RPCPublicKey{},
		IP_Whitelist:   []multiaddr.Multiaddr{},
	},
}

type Config struct {
	PrivKey crypto.PrivKey

	PublicAddress    multiaddr.Multiaddr
	SeedAddresses    []multiaddr.Multiaddr
	ConnectAddresses []multiaddr.Multiaddr
	ListenAddresses  []multiaddr.Multiaddr

	Params *cfg.Params

	RPCConfig *rpc.RPCConfig
}
