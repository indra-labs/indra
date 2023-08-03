package forward

import (
	"github.com/indra-labs/indra/pkg/codec"
	"github.com/indra-labs/indra/pkg/codec/ont"
	"github.com/indra-labs/indra/pkg/codec/reg"
	"github.com/indra-labs/indra/pkg/util/multi"
	"github.com/multiformats/go-multiaddr"
	"math/rand"
	"net"
	"net/netip"
	"reflect"
	"testing"

	"github.com/indra-labs/indra/pkg/crypto/nonce"
)

func TestOnions_Forward(t *testing.T) {
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
		s := ont.Encode(on)
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
		var cf *Forward
		if cf, ok = onr.(*Forward); !ok {
			t.Error("did not unwrap expected type expected *Reverse got",
				reflect.TypeOf(onr))
			t.FailNow()
		}
		if cf.Multiaddr.String() != ma.String() {
			t.Error("reverse Addresses did not unwrap correctly")
			t.FailNow()
		}
	}
}
