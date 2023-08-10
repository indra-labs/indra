// Package ads provides a bundle for peer information advertisement types and initial generation of them, and deriving a peer node data structure from the ad set received over the gossip network.
package ads

import (
	"errors"
	"git.indra-labs.org/dev/ind/pkg/cert"
	"git.indra-labs.org/dev/ind/pkg/codec/ad"
	"git.indra-labs.org/dev/ind/pkg/codec/ad/addresses"
	"git.indra-labs.org/dev/ind/pkg/codec/ad/load"
	"git.indra-labs.org/dev/ind/pkg/codec/ad/peer"
	services2 "git.indra-labs.org/dev/ind/pkg/codec/ad/services"
	"git.indra-labs.org/dev/ind/pkg/crypto"
	"git.indra-labs.org/dev/ind/pkg/crypto/nonce"
	"git.indra-labs.org/dev/ind/pkg/engine/node"
	"git.indra-labs.org/dev/ind/pkg/engine/payments"
	"git.indra-labs.org/dev/ind/pkg/engine/services"
	log2 "git.indra-labs.org/dev/ind/pkg/proc/log"
	"git.indra-labs.org/dev/ind/pkg/util/multikey"
	"github.com/multiformats/go-multiaddr"
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
	Peer     *peer.Ad
	Address  *addresses.Ad
	Services *services2.Ad
	Load     *load.Ad
}

func (na *NodeAds) GetAsCerts() (ads []cert.Act) {
	return []cert.Act{na.Address, na.Load, na.Peer, na.Services}
}

// GetMultiaddrs returns a node's listener addresses.
func GetMultiaddrs(n *node.Node) (ma []multiaddr.Multiaddr, e error) {
	for _, aa := range n.Addresses {
		ma = append(ma, multikey.AddKeyToMultiaddr(aa, n.Identity.Pub))
	}
	return
}

func GetServices(n *node.Node) (svcs []services2.Service) {
	for i := range n.Services {
		svcs = append(svcs, services2.Service{
			Port:      n.Services[i].Port,
			RelayRate: n.Services[i].RelayRate,
		})
	}
	return
}

// GenerateAds takes a node.Node and creates the NodeAds matching it.
func GenerateAds(n *node.Node, ld byte) (na *NodeAds, e error) {
	expiry := time.Now().Add(DefaultAdExpiry)
	s := GetServices(n)
	var ma []multiaddr.Multiaddr
	ma, e = GetMultiaddrs(n)
	if fails(e) {
		return
	}
	na = &NodeAds{
		Peer: &peer.Ad{
			Ad: ad.Ad{
				ID:     nonce.NewID(),
				Key:    n.Identity.Pub,
				Expiry: expiry,
			},
			RelayRate: n.RelayRate,
		},
		Address: &addresses.Ad{
			Ad: ad.Ad{
				ID:     nonce.NewID(),
				Key:    n.Identity.Pub,
				Expiry: expiry,
			},
			Addresses: ma,
		},
		Services: &services2.Ad{
			Ad: ad.Ad{
				ID:     nonce.NewID(),
				Key:    n.Identity.Pub,
				Expiry: expiry,
			},
			Services: s,
		},
		Load: &load.Ad{
			Ad: ad.Ad{
				ID:     nonce.NewID(),
				Key:    n.Identity.Pub,
				Expiry: time.Now().Add(time.Minute * 10),
			},
			Load: ld,
		},
	}
	return
}

// ErrNilNodeAds indicates that the NodeAds provided was nil.
const ErrNilNodeAds = "cannot process nil NodeAds"

// NodeFromAds generates a node.Node from a NodeAds. Used by clients to create
// models for peers they have sessions with.
func NodeFromAds(a *NodeAds) (n *node.Node, e error) {
	if a == nil ||
		a.Services == nil || a.Load == nil ||
		a.Address == nil || a.Peer == nil {
		return n, errors.New(ErrNilNodeAds)
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
		ID:        nonce.NewID(),
		Addresses: a.Address.Addresses,
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
