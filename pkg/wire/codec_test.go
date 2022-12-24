package wire

import (
	"math/rand"
	"net"
	"net/netip"
	"reflect"
	"testing"
	"time"

	"github.com/Indra-Labs/indra/pkg/key/prv"
	"github.com/Indra-Labs/indra/pkg/key/pub"
	"github.com/Indra-Labs/indra/pkg/nonce"
	"github.com/Indra-Labs/indra/pkg/sha256"
	"github.com/Indra-Labs/indra/pkg/slice"
	"github.com/Indra-Labs/indra/pkg/testutils"
	"github.com/Indra-Labs/indra/pkg/types"
	"github.com/Indra-Labs/indra/pkg/wire/cipher"
	"github.com/Indra-Labs/indra/pkg/wire/confirmation"
	"github.com/Indra-Labs/indra/pkg/wire/delay"
	"github.com/Indra-Labs/indra/pkg/wire/exit"
	"github.com/Indra-Labs/indra/pkg/wire/forward"
	log2 "github.com/cybriq/proc/pkg/log"
)

func TestOnionSkins_Cipher(t *testing.T) {
	log2.CodeLoc = true
	var e error
	hdrP, pldP := GetTwoPrvKeys(t)
	hdr, pld := pub.Derive(hdrP), pub.Derive(pldP)
	on := OnionSkins{}.
		Cipher(hdr, pld).
		Assemble()
	onb := EncodeOnion(on)
	c := slice.NewCursor()
	var onc types.Onion
	if onc, e = PeelOnion(onb, c); check(e) {
		t.FailNow()
	}
	var ci *cipher.OnionSkin
	var ok bool
	if ci, ok = onc.(*cipher.OnionSkin); !ok {
		t.Error("did not unwrap expected type")
		t.FailNow()
	}
	if !ci.Header.Equals(hdr) {
		t.Error("header key did not unwrap correctly")
		t.FailNow()
	}
	if !ci.Payload.Equals(pld) {
		t.Error("payload key did not unwrap correctly")
		t.FailNow()
	}
}

func TestOnionSkins_Confirmation(t *testing.T) {
	log2.CodeLoc = true
	var e error
	n := nonce.NewID()
	on := OnionSkins{}.
		Confirmation(n).
		Assemble()
	onb := EncodeOnion(on)
	c := slice.NewCursor()
	var oncn types.Onion
	if oncn, e = PeelOnion(onb, c); check(e) {
		t.FailNow()
	}
	var cf *confirmation.OnionSkin
	var ok bool
	if cf, ok = oncn.(*confirmation.OnionSkin); !ok {
		t.Error("did not unwrap expected type")
		t.FailNow()
	}
	if cf.ID != n {
		t.Error("confirmation ID did not unwrap correctly")
		t.FailNow()
	}
}
func TestOnionSkins_Delay(t *testing.T) {
	log2.CodeLoc = true
	var e error
	del := time.Duration(rand.Uint64())
	on := OnionSkins{}.
		Delay(del).
		Assemble()
	onb := EncodeOnion(on)
	c := slice.NewCursor()
	var ond types.Onion
	if ond, e = PeelOnion(onb, c); check(e) {
		t.FailNow()
	}
	var dl *delay.OnionSkin
	var ok bool
	if dl, ok = ond.(*delay.OnionSkin); !ok {
		t.Error("did not unwrap expected type")
		t.FailNow()
	}
	if dl.Duration != del {
		t.Error("delay duration did not unwrap correctly")
		t.FailNow()
	}
}

func TestOnionSkins_Exit(t *testing.T) {
	log2.CodeLoc = true
	var e error
	var msg slice.Bytes
	prvs, pubs := GetCipherSet(t)
	ciphers := GenCiphers(prvs, pubs)
	var hash sha256.Hash
	if msg, hash, e = testutils.GenerateTestMessage(512); check(e) {
		t.Error(e)
		t.FailNow()
	}
	p := uint16(rand.Uint32())
	on := OnionSkins{}.
		Exit(p, prvs, pubs, msg).
		Assemble()
	onb := EncodeOnion(on)
	c := slice.NewCursor()
	var onex types.Onion
	if onex, e = PeelOnion(onb, c); check(e) {
		t.FailNow()
	}
	var ex *exit.OnionSkin
	var ok bool
	if ex, ok = onex.(*exit.OnionSkin); !ok {
		t.Error("did not unwrap expected type")
		t.FailNow()
	}
	if ex.Port != p {
		t.Error("exit port did not unwrap correctly")
		t.FailNow()
	}
	for i := range ex.Ciphers {
		if ex.Ciphers[i] != ciphers[i] {
			t.Errorf("cipher %d did not unwrap correctly", i)
			t.FailNow()
		}
	}
	plH := sha256.Single(ex.Bytes)
	if plH != hash {
		t.Errorf("exit message did not unwrap correctly")
		t.FailNow()
	}
}

func TestOnionSkins_Forward(t *testing.T) {
	log2.CodeLoc = true
	var e error
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
		on := OnionSkins{}.
			Forward(ap).
			Assemble()
		onb := EncodeOnion(on)
		c := slice.NewCursor()
		var onf types.Onion
		if onf, e = PeelOnion(onb, c); check(e) {
			t.FailNow()
		}
		var cf *forward.OnionSkin
		if cf, ok = onf.(*forward.OnionSkin); !ok {
			t.Error("did not unwrap expected type", reflect.TypeOf(onf))
			t.FailNow()
		}
		_ = cf
		// if !cf.IP.Equal(ip) {
		// 	t.Error("forward IP did not unwrap correctly")
		// 	t.FailNow()
		// }
	}
}

func TestOnionSkins_Message(t *testing.T) {

}

func TestOnionSkins_Purchase(t *testing.T) {

}

func TestOnionSkins_Reply(t *testing.T) {

}

func TestOnionSkins_Response(t *testing.T) {

}

func TestOnionSkins_Session(t *testing.T) {

}

func TestOnionSkins_Token(t *testing.T) {

}

func GetTwoPrvKeys(t *testing.T) (prv1, prv2 *prv.Key) {
	var e error
	if prv1, e = prv.GenerateKey(); check(e) {
		t.FailNow()
	}
	if prv2, e = prv.GenerateKey(); check(e) {
		t.FailNow()
	}
	return
}

func GetCipherSet(t *testing.T) (prvs [3]*prv.Key, pubs [3]*pub.Key) {
	for i := range prvs {
		prv1, prv2 := GetTwoPrvKeys(t)
		prvs[i] = prv1
		pubs[i] = pub.Derive(prv2)
	}
	return
}
