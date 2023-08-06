// Package multi provides several functions for working with multiaddr.Multiaddr and netip.AddrPort types, including public key p2p identifiers.
package multi

import (
	"errors"
	"fmt"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
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

// AddrToBytes takes a multiaddr, strips the ip4/ip6/dns address and port out of
// it, and recombines it to generate a binary form of the essential network
// address element of a multiaddr.
//
// Note that as mentioned elsewhere, Indra only uses TCP because of the
// incomplete multi-platform support for QUIC, and a tendency of aggressive
// network filtering firewalls to prejudicially block UDP traffic.
func AddrToBytes(ma multiaddr.Multiaddr, defaultPort uint16) (b []byte,
	e error) {

	var reassembled []multiaddr.Multiaddr
	pieces := multiaddr.Split(ma)

	var portFound, firstAddr bool
	for _, v := range pieces {
		switch {
		case v.Protocols()[0].Code == multiaddr.P_TCP:

			// This should be the port
			reassembled = append(reassembled, v)
			portFound = true

		case v.Protocols()[0].Code == multiaddr.P_DNS ||
			v.Protocols()[0].Code == multiaddr.P_IP4 ||
			v.Protocols()[0].Code == multiaddr.P_IP6:

			// This is the address
			reassembled = append(reassembled, v)
			firstAddr = true
		}
		if portFound && firstAddr {
			// we have all we need, the multiaddr should not have more than one
			// of each type.
			break
		}
	}

	// If there isn't a port, add the default port for the network.
	if !portFound {

		m, e := multiaddr.NewMultiaddr(fmt.Sprintf("/tcp/%d",
			defaultPort))

		if fails(e) {
			panic(e)
		}

		reassembled = append(reassembled, m)
	}

	// If there is no address return this as an error.
	if !firstAddr {
		e = fmt.Errorf("multiaddr %v does not have an address",
			ma)
		return
	}

	// We can now assume we have two members in `reassembled`
	return reassembled[0].Encapsulate(reassembled[1]).MarshalBinary()
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
