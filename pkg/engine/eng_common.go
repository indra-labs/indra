package engine

import (
	"net"
	"net/netip"
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
