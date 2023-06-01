package onions

import (
	"github.com/indra-labs/indra/pkg/engine/onions/reg"
	"math/rand"
	"net"
	"net/netip"
	"reflect"
	"testing"

	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/engine/coding"
)

func TestOnions_Reverse(t *testing.T) {
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
			t.Error("unable to get netip.Addr")
			t.FailNow()
		}
		port := uint16(rand.Uint32())
		ap := netip.AddrPortFrom(adr, port)
		on := Skins{}.
			Reverse(&ap).
			Assemble()
		s := Encode(on)
		s.SetCursor(0)
		var onr coding.Codec
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
		if cf.AddrPort.String() != ap.String() {
			log.I.S(cf.AddrPort, ap)
			t.Error("reverse AddrPort did not unwrap correctly")
			t.FailNow()
		}
	}
}
