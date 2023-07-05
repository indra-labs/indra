package forward

import (
	"github.com/indra-labs/indra/pkg/codec"
	"github.com/indra-labs/indra/pkg/codec/ont"
	"github.com/indra-labs/indra/pkg/codec/reg"
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
		on := ont.Assemble([]ont.Onion{
			NewForward(&ap),
		})
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
		if cf.AddrPort.String() != ap.String() {
			log.I.S(cf.AddrPort, ap)
			t.Error("reverse Addresses did not unwrap correctly")
			t.FailNow()
		}
	}
}
