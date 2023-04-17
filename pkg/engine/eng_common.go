package engine

import (
	"net"
	"net/netip"
	
	"git-indra.lan/indra-labs/indra"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	fails = log.E.Chk
)

func GetNetworkFromAddrPort(addr string) (nw string, u *net.UDPAddr,
	e error) {
	
	nw = "udp"
	var ap netip.AddrPort
	if ap, e = netip.ParseAddrPort(addr); fails(e) {
		return
	}
	u = &net.UDPAddr{IP: net.ParseIP(ap.Addr().String()), Port: int(ap.Port())}
	if u.IP.To4() != nil {
		nw = "udp4"
	}
	return
}
