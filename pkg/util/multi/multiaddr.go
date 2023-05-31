package multi

import (
	"github.com/indra-labs/indra"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/multiformats/go-multiaddr"
	"net/netip"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	fails = log.E.Chk
)

func AddrToAddrPort(ma multiaddr.Multiaddr) (ap netip.AddrPort, e error) {
	var addrStr string
	if addrStr, e = ma.ValueForProtocol(multiaddr.P_IP4); fails(e) {
		if addrStr, e = ma.ValueForProtocol(multiaddr.P_IP6); fails(e) {
			return
		}
	}
	var portStr string
	if portStr, e = ma.ValueForProtocol(multiaddr.P_TCP); fails(e) {
		return
	}
	if ap, e = netip.ParseAddrPort(addrStr + ":" + portStr); fails(e) {
		return
	}
	return
}
