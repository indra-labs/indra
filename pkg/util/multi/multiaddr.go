// Package multi provides several functions for working with multiaddr.Multiaddr and netip.AddrPort types, including public key p2p identifiers.
package multi

import (
	"errors"
	"fmt"
	"github.com/indra-labs/indra/pkg/crypto"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"net/netip"
)

var (
	log   = log2.GetLogger()
	fails = log.E.Chk
)

func AddrToAddrPort(ma multiaddr.Multiaddr) (ap netip.AddrPort, e error) {
	if ma == nil {
		e = errors.New("nil multiaddr.Multiaddr")
		return
	}
	var addrStr string
	var is6 bool
	if addrStr, e = ma.ValueForProtocol(multiaddr.P_IP4); e != nil {
		if addrStr, e = ma.ValueForProtocol(multiaddr.P_IP6); fails(e) {
			return
		} else {
			is6 = true
		}
	}
	var portStr string
	if portStr, e = ma.ValueForProtocol(multiaddr.P_TCP); fails(e) {
		return
	}
	if is6 {
		addrStr = "[" + addrStr + "]"
	}
	if ap, e = netip.ParseAddrPort(addrStr + ":" + portStr); fails(e) {
		return
	}
	return
}

func AddrFromAddrPort(ap netip.AddrPort) (ma multiaddr.Multiaddr, e error) {
	var ipv string
	if ap.Addr().Is6() {
		ipv = "ip6"
	} else {
		ipv = "ip4"
	}
	if ma, e = multiaddr.NewMultiaddr(fmt.Sprintf("/%s/%s/tcp/%d",
		ipv, ap.Addr().String(), ap.Port())); fails(e) {
		return
	}
	return
}

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
