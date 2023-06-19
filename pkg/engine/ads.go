package engine

import (
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/engine/node"
	"github.com/indra-labs/indra/pkg/onions/adaddress"
	"github.com/indra-labs/indra/pkg/onions/adpeer"
	"github.com/indra-labs/indra/pkg/onions/adproto"
	"github.com/indra-labs/indra/pkg/onions/adservices"
	"github.com/indra-labs/indra/pkg/util/multi"
	"github.com/multiformats/go-multiaddr"
	"time"
)

const DefaultAdExpiry = time.Hour * 24 * 7 // one week

type NodeAds struct {
	Peer     adpeer.Ad
	Address  adaddress.Ad
	Services adservices.Ad
}

func GetMultiaddr(n *node.Node) (ma multiaddr.Multiaddr, e error) {
	if ma, e = multi.AddrFromAddrPort(*n.AddrPort); fails(e) {
		return
	}
	ma = multi.AddKeyToMultiaddr(ma, n.Identity.Pub)
	return
}

func GenerateAds(n *node.Node) (na *NodeAds, e error) {
	expiry := time.Now().Add(DefaultAdExpiry)
	var svcs []adservices.Service
	for i := range n.Services {
		svcs = append(svcs, adservices.Service{
			Port:      n.Services[i].Port,
			RelayRate: uint32(n.Services[i].RelayRate),
		})
	}
	var ma multiaddr.Multiaddr
	if ma, e = multi.AddrFromAddrPort(*n.AddrPort); fails(e) {
		return
	}
	ma = multi.AddKeyToMultiaddr(ma, n.Identity.Pub)
	na = &NodeAds{
		Peer: adpeer.Ad{
			Ad: adproto.Ad{
				ID:     nonce.NewID(),
				Key:    n.Identity.Pub,
				Expiry: expiry,
			},
			RelayRate: n.RelayRate,
		},
		Address: adaddress.Ad{
			Ad: adproto.Ad{
				ID:     nonce.NewID(),
				Key:    n.Identity.Pub,
				Expiry: expiry,
			},
			Addr: ma,
		},
		Services: adservices.Ad{
			Ad: adproto.Ad{
				ID:     nonce.NewID(),
				Key:    n.Identity.Pub,
				Expiry: expiry,
			},
			Services: svcs,
		},
	}
	return
}
