package ads

import (
	"errors"
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/engine/node"
	"github.com/indra-labs/indra/pkg/engine/payments"
	"github.com/indra-labs/indra/pkg/engine/services"
	"github.com/indra-labs/indra/pkg/onions/adaddress"
	"github.com/indra-labs/indra/pkg/onions/adload"
	"github.com/indra-labs/indra/pkg/onions/adpeer"
	"github.com/indra-labs/indra/pkg/onions/adproto"
	"github.com/indra-labs/indra/pkg/onions/adservices"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/multi"
	"github.com/multiformats/go-multiaddr"
	"net/netip"
	"time"
)

var (
	log   = log2.GetLogger()
	fails = log.E.Chk
)

// DefaultAdExpiry is the base expiry duration
//
// todo: 1 week? Should this be shorter?
const DefaultAdExpiry = time.Hour * 24 * 7 // one week

// NodeAds are all the ads associated with a peer.
//
// Some are longer lived than others, mostly Peer and Address ads will last a
// long time but services might change more often and load will be updated
// whenever it dramatically changes or every few minutes.
type NodeAds struct {
	Peer     *adpeer.Ad
	Address  *adaddress.Ad
	Services *adservices.Ad
	Load     *adload.Ad
}

// GetMultiaddr returns a node's listener address.
func GetMultiaddr(n *node.Node) (ma multiaddr.Multiaddr, e error) {
	if ma, e = multi.AddrFromAddrPort(*n.AddrPort); fails(e) {
		return
	}
	ma = multi.AddKeyToMultiaddr(ma, n.Identity.Pub)
	return
}

func GenerateAds(n *node.Node, load byte) (na *NodeAds, e error) {
	expiry := time.Now().Add(DefaultAdExpiry)
	var svcs []adservices.Service
	for i := range n.Services {
		svcs = append(svcs, adservices.Service{
			Port:      n.Services[i].Port,
			RelayRate: n.Services[i].RelayRate,
		})
	}
	var ma multiaddr.Multiaddr
	if ma, e = GetMultiaddr(n); fails(e) {
		return
	}
	na = &NodeAds{
		Peer: &adpeer.Ad{
			Ad: adproto.Ad{
				ID:     nonce.NewID(),
				Key:    n.Identity.Pub,
				Expiry: expiry,
			},
			RelayRate: n.RelayRate,
		},
		Address: &adaddress.Ad{
			Ad: adproto.Ad{
				ID:     nonce.NewID(),
				Key:    n.Identity.Pub,
				Expiry: expiry,
			},
			Addrs: ma,
		},
		Services: &adservices.Ad{
			Ad: adproto.Ad{
				ID:     nonce.NewID(),
				Key:    n.Identity.Pub,
				Expiry: expiry,
			},
			Services: svcs,
		},
		Load: &adload.Ad{
			Ad: adproto.Ad{
				ID:     nonce.NewID(),
				Key:    n.Identity.Pub,
				Expiry: time.Now().Add(time.Minute * 10),
			},
			Load: load,
		},
	}
	return
}

const ErrNilNodeAds = "cannot process nil NodeAds"

func NodeFromAds(a *NodeAds) (n *node.Node, e error) {
	if a == nil ||
		a.Services == nil || a.Load == nil ||
		a.Address == nil || a.Peer == nil {
		return n, errors.New(ErrNilNodeAds)
	}
	var ap netip.AddrPort
	if ap, e = multi.AddrToAddrPort(a.Address.Addrs); fails(e) {
		return
	}
	var svcs services.Services
	for i := range a.Services.Services {
		svcs = append(svcs, &services.Service{
			Port:      a.Services.Services[i].Port,
			RelayRate: a.Services.Services[i].RelayRate,
			Transport: nil, // todo: wen making?
		})
	}
	n = &node.Node{
		ID:       nonce.NewID(),
		AddrPort: &ap,
		Identity: &crypto.Keys{
			Pub:   a.Address.Key,
			Bytes: a.Address.Key.ToBytes(),
		},
		RelayRate: a.Peer.RelayRate,
		Services:  svcs,
		Load:      a.Load.Load,
		PayChan:   make(payments.PayChan, node.PaymentChanBuffers), // todo: other end stuff
		Transport: nil,                                             // this is populated when we dial it.
	}
	return
}
