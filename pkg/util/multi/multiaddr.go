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

// AddrToBytes takes a multiaddr, strips the ip4/ip6/dns address and port out of
// it, and recombines it to generate a binary form of the essential network
// address element of a multiaddr.
//
// Note that as mentioned elsewhere, Indra only uses TCP because of the
// incomplete multi-platform support for QUIC, and a tendency of aggressive
// network filtering firewalls to prejudicially block UDP traffic.
func AddrToBytes(ma multiaddr.Multiaddr, defaultPort uint16) (b []byte,
	e error) {

	// First, read out the values encoded for each of the relevant IP and port
	// in the value.
	var ip, port string
	ip, e = ma.ValueForProtocol(multiaddr.P_IP4)
	if e != nil || ip == "" {
		ip, e = ma.ValueForProtocol(multiaddr.P_IP6)
		if e != nil || ip == "" {
			ip, e = ma.ValueForProtocol(multiaddr.P_DNS)
		}
	}
	if fails(e) || ip == "" {
		// There must be DNS, ip4 or ip6 addresses for this field.
		return
	}
	port, e = ma.ValueForProtocol(multiaddr.P_TCP)
	if fails(e) {
		return
	}
	// If the port is missing, replace it with the defaultPort.
	if port == "" {
		var pma multiaddr.Multiaddr
		pma, e = multiaddr.NewMultiaddr(fmt.Sprintf("/tcp/%d",
			defaultPort))
		if fails(e) {
			// This shouldn't happen.
			return
		}
		port, _ = pma.ValueForProtocol(multiaddr.P_TCP)
	}

	// Assemble the address and port, and then return the binary form.
	var mip, port2 multiaddr.Multiaddr
	mip, e = multiaddr.NewMultiaddr(ip)
	port2, e = multiaddr.NewMultiaddr(port)
	return mip.Encapsulate(port2).MarshalBinary()
}

// BytesToMultiaddr will usually be operating on an encoded value produced by
// AddrToBytes, meaning it will only encode ip4/ip6/dns and tcp port.
//
// It will of course decode any valid protocols encoded in the byte slice, just
// that the data currently is never encoded into binary with any other data
// present, mainly because the private key is a distinct identity and is used in
// Indra for a relay, clients, and hidden services and for now there isn't a
// multiaddr implementation for indra Hidden Services (there will be!).
func BytesToMultiaddr(b []byte) (ma multiaddr.Multiaddr,
	e error) {

	return multiaddr.NewMultiaddrBytes(b)
}
