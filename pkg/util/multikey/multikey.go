package multikey

import (
	"github.com/indra-labs/indra/pkg/crypto"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

var (
	log   = log2.GetLogger()
	fails = log.E.Chk
)

func AddKeyToMultiaddr(in multiaddr.Multiaddr, pub *crypto.Pub) (ma multiaddr.Multiaddr) {
	var pid peer.ID
	var e error
	if pid, e = peer.IDFromPublicKey(pub); fails(e) {
		return
	}
	var k multiaddr.Multiaddr
	if k, e = multiaddr.NewMultiaddr("/p2p/" + pid.String()); fails(e) {
		return
	}
	ma = in.Encapsulate(k)
	return
}
