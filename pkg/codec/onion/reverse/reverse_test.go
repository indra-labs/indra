package reverse

import (
	"git.indra-labs.org/dev/ind"
	"git.indra-labs.org/dev/ind/pkg/codec"
	"git.indra-labs.org/dev/ind/pkg/codec/ont"
	"git.indra-labs.org/dev/ind/pkg/codec/reg"
	log2 "git.indra-labs.org/dev/ind/pkg/proc/log"
	"git.indra-labs.org/dev/ind/pkg/util/multi"
	"github.com/multiformats/go-multiaddr"
	"math/rand"
	"net"
	"net/netip"
	"reflect"
	"testing"

	"git.indra-labs.org/dev/ind/pkg/crypto/nonce"
)

func TestOnions_Reverse(t *testing.T) {
	if indra.CI == "false" {
		log2.SetLogLevel(log2.Trace)
	}
	ipSizes := []int{net.IPv4len, net.IPv6len}
	for i := range ipSizes {
		n := nonce.New()
		ip := net.IP(n[:ipSizes[i]])
		var adr netip.Addr
		if ipSizes[i] == net.IPv4len {
			ip = ip.To4()
		}
		if ipSizes[i] == net.IPv6len {
			ip = ip.To16()
		}
		var ok bool
		if adr, ok = netip.AddrFromSlice(ip); !ok {
			t.Error("unable to get netip.Addrs")
			t.FailNow()
		}
		port := uint16(rand.Uint32())
		ap := netip.AddrPortFrom(adr, port)
		var ma multiaddr.Multiaddr
		var e error
		if ma, e = multi.AddrFromAddrPort(ap); fails(e) {
			t.FailNow()
		}
		on := ont.Assemble([]ont.Onion{New(ma)})
		s := codec.Encode(on)
		s.SetCursor(0)
		var onr codec.Codec
		if onr = reg.Recognise(s); onr == nil {
			t.Error("did not unwrap")
			t.FailNow()
		}
		if e := onr.Decode(s); fails(e) {
			t.Error("did not decode")
			t.FailNow()

		}
		var cf *Reverse
		if cf, ok = onr.(*Reverse); !ok {
			t.Error("did not unwrap expected type expected *Return got",
				reflect.TypeOf(onr))
			t.FailNow()
		}
		if cf.Multiaddr.String() != ma.String() {
			log.I.S(cf.Multiaddr, ap)
			t.Error("reverse Addresses did not unwrap correctly")
			t.FailNow()
		}
	}
}
