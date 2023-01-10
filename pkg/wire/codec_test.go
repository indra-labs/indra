package wire

import (
	"math/rand"
	"net"
	"net/netip"
	"reflect"
	"testing"
	"time"

	"github.com/indra-labs/indra/pkg/key/prv"
	"github.com/indra-labs/indra/pkg/key/pub"
	"github.com/indra-labs/indra/pkg/nonce"
	"github.com/indra-labs/indra/pkg/sha256"
	"github.com/indra-labs/indra/pkg/slice"
	"github.com/indra-labs/indra/pkg/testutils"
	"github.com/indra-labs/indra/pkg/types"
	"github.com/indra-labs/indra/pkg/wire/cipher"
	"github.com/indra-labs/indra/pkg/wire/confirm"
	"github.com/indra-labs/indra/pkg/wire/delay"
	"github.com/indra-labs/indra/pkg/wire/exit"
	"github.com/indra-labs/indra/pkg/wire/forward"
	"github.com/indra-labs/indra/pkg/wire/layer"
	"github.com/indra-labs/indra/pkg/wire/response"
	"github.com/indra-labs/indra/pkg/wire/reverse"
	"github.com/indra-labs/indra/pkg/wire/session"
	"github.com/indra-labs/indra/pkg/wire/token"
)

func TestOnionSkins_Cipher(t *testing.T) {

	var e error
	hdrP, pldP := GetTwoPrvKeys(t)
	// hdr, pld := pub.Derive(hdrP), pub.Derive(pldP)
	on := OnionSkins{}.
		Cipher(hdrP, pldP).
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
	if !ci.Header.Key.Equals(&hdrP.Key) {
		t.Error("header key did not unwrap correctly")
		t.FailNow()
	}
	if !ci.Payload.Key.Equals(&pldP.Key) {
		t.Error("payload key did not unwrap correctly")
		t.FailNow()
	}
}

func TestOnionSkins_Confirmation(t *testing.T) {

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
	var cf *confirm.OnionSkin
	var ok bool
	if cf, ok = oncn.(*confirm.OnionSkin); !ok {
		t.Error("did not unwrap expected type")
		t.FailNow()
	}
	if cf.ID != n {
		t.Error("confirmation ID did not unwrap correctly")
		t.FailNow()
	}
}

func TestOnionSkins_Delay(t *testing.T) {

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

	var e error
	prvs, pubs := GetCipherSet(t)
	ciphers := GenCiphers(prvs, pubs)
	var msg slice.Bytes
	var hash sha256.Hash
	if msg, hash, e = testutils.GenerateTestMessage(512); check(e) {
		t.Error(e)
		t.FailNow()
	}
	n3 := Gen3Nonces()
	p := uint16(rand.Uint32())
	on := OnionSkins{}.
		Exit(p, prvs, pubs, n3, msg).
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
	for i := range ex.Nonces {
		if ex.Nonces[i] != n3[i] {
			t.Errorf("nonce %d did not unwrap correctly", i)
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
			Forward(&ap).
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
		if cf.AddrPort.String() != ap.String() {
			log.I.S(cf.AddrPort, ap)
			t.Error("forward AddrPort did not unwrap correctly")
			t.FailNow()
		}
	}
}

func TestOnionSkins_Layer(t *testing.T) {

	var e error
	n := nonce.NewID()
	n1 := nonce.New()
	prv1, prv2 := GetTwoPrvKeys(t)
	pub1 := pub.Derive(prv1)
	on := OnionSkins{}.
		OnionSkin(pub1, prv2, n1).
		Confirmation(n).
		Assemble()
	onb := EncodeOnion(on)
	c := slice.NewCursor()
	var onos, onc types.Onion
	if onos, e = PeelOnion(onb, c); check(e) {
		t.Error(e)
		t.FailNow()
	}
	os := &layer.OnionSkin{}
	var ok bool
	if os, ok = onos.(*layer.OnionSkin); !ok {
		t.Error("did not unwrap expected type")
		t.FailNow()
	}
	os.Decrypt(prv1, onb, c)
	// unwrap the confirmation
	if onc, e = PeelOnion(onb, c); check(e) {
		t.Error(e)
		t.FailNow()
	}
	oc := &confirm.OnionSkin{}
	if oc, ok = onc.(*confirm.OnionSkin); !ok {
		t.Error("did not unwrap expected type")
		t.FailNow()
	}
	if oc.ID != n {
		t.Error("did not recover the confirmation nonce")
		t.FailNow()
	}
}

func TestOnionSkins_Reply(t *testing.T) {

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
			Reverse(&ap).
			Assemble()
		onb := EncodeOnion(on)
		c := slice.NewCursor()
		var onf types.Onion
		if onf, e = PeelOnion(onb, c); check(e) {
			t.FailNow()
		}
		var cf *reverse.OnionSkin
		if cf, ok = onf.(*reverse.OnionSkin); !ok {
			t.Error("did not unwrap expected type", reflect.TypeOf(onf))
			t.FailNow()
		}
		if cf.AddrPort.String() != ap.String() {
			log.I.S(cf.AddrPort, ap)
			t.Error("reply AddrPort did not unwrap correctly")
			t.FailNow()
		}
	}
}

func TestOnionSkins_Response(t *testing.T) {

	var e error
	var msg slice.Bytes
	var hash sha256.Hash
	if msg, hash, e = testutils.GenerateTestMessage(10000); check(e) {
		t.Error(e)
		t.FailNow()
	}
	on := OnionSkins{}.
		Response(hash, msg).
		Assemble()
	onb := EncodeOnion(on)
	c := slice.NewCursor()
	var onex types.Onion
	if onex, e = PeelOnion(onb, c); check(e) {
		t.FailNow()
	}
	ex := &response.OnionSkin{}
	var ok bool
	if ex, ok = onex.(*response.OnionSkin); !ok {
		t.Error("did not unwrap expected type")
		t.FailNow()
	}
	plH := sha256.Single(ex.Bytes)
	if plH != hash {
		t.Errorf("exit message did not unwrap correctly")
		t.FailNow()
	}

}

func TestOnionSkins_Session(t *testing.T) {

	var e error
	hdrP, pldP := GetTwoPrvKeys(t)
	hdr, pld := pub.Derive(hdrP), pub.Derive(pldP)
	on := OnionSkins{}.
		Session(hdr, pld).
		Assemble()
	onb := EncodeOnion(on)
	c := slice.NewCursor()
	var oncn types.Onion
	if oncn, e = PeelOnion(onb, c); check(e) {
		t.FailNow()
	}
	var cf *session.OnionSkin
	var ok bool
	if cf, ok = oncn.(*session.OnionSkin); !ok {
		t.Error("did not unwrap expected type", reflect.TypeOf(oncn))
		t.FailNow()
	}
	if !cf.HeaderKey.Equals(hdr) {
		t.Error("header key mismatch")
		t.FailNow()
	}
	if !cf.PayloadKey.Equals(pld) {
		t.Error("payload key mismatch")
		t.FailNow()
	}

}

func TestOnionSkins_Token(t *testing.T) {

	var e error
	ni := nonce.NewID()
	n := sha256.Single(ni[:])
	on := OnionSkins{}.
		Token(n).
		Assemble()
	onb := EncodeOnion(on)
	c := slice.NewCursor()
	var oncn types.Onion
	if oncn, e = PeelOnion(onb, c); check(e) {
		t.FailNow()
	}
	var cf *token.OnionSkin
	var ok bool
	if cf, ok = oncn.(*token.OnionSkin); !ok {
		t.Error("did not unwrap expected type", reflect.TypeOf(oncn))
		t.FailNow()
	}
	if sha256.Hash(*cf) != n {
		log.I.S(n, cf)
		t.Error("confirmation ID did not unwrap correctly")
		t.FailNow()
	}
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
