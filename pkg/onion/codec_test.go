package onion

import (
	"math/rand"
	"net"
	"net/netip"
	"reflect"
	"testing"
	"time"

	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/confirm"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/crypt"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/delay"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/exit"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/forward"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/response"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/reverse"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/session"
	"git-indra.lan/indra-labs/indra/pkg/types"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
	"git-indra.lan/indra-labs/indra/pkg/util/tests"
)

func TestOnionSkins_Cipher(t *testing.T) {

	var e error
	sess := session.New(1)
	on := Skins{}.
		Session(sess).
		Assemble()
	onb := Encode(on)
	c := slice.NewCursor()
	var onc types.Onion
	if onc, e = Peel(onb, c); check(e) {
		t.FailNow()
	}
	var ci *session.Layer
	var ok bool
	if ci, ok = onc.(*session.Layer); !ok {
		t.Error("did not unwrap expected type")
		t.FailNow()
	}
	if !ci.Header.Key.Equals(&sess.Header.Key) {
		t.Error("header key did not unwrap correctly")
		t.FailNow()
	}
	if !ci.Payload.Key.Equals(&sess.Payload.Key) {
		t.Error("payload key did not unwrap correctly")
		t.FailNow()
	}
}

func TestOnionSkins_Confirmation(t *testing.T) {

	var e error
	n := nonce.NewID()
	on := Skins{}.
		Confirmation(n).
		Assemble()
	onb := Encode(on)
	c := slice.NewCursor()
	var oncn types.Onion
	if oncn, e = Peel(onb, c); check(e) {
		t.FailNow()
	}
	var cf *confirm.Layer
	var ok bool
	if cf, ok = oncn.(*confirm.Layer); !ok {
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
	on := Skins{}.
		Delay(del).
		Assemble()
	onb := Encode(on)
	c := slice.NewCursor()
	var ond types.Onion
	if ond, e = Peel(onb, c); check(e) {
		t.FailNow()
	}
	var dl *delay.Layer
	var ok bool
	if dl, ok = ond.(*delay.Layer); !ok {
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
	if msg, hash, e = tests.GenMessage(512, ""); check(e) {
		t.Error(e)
		t.FailNow()
	}
	n3 := Gen3Nonces()
	p := uint16(rand.Uint32())
	id := nonce.NewID()
	on := Skins{}.
		Exit(p, prvs, pubs, n3, id, msg).
		Assemble()
	onb := Encode(on)
	c := slice.NewCursor()
	var onex types.Onion
	if onex, e = Peel(onb, c); check(e) {
		t.FailNow()
	}
	var ex *exit.Layer
	var ok bool
	if ex, ok = onex.(*exit.Layer); !ok {
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
	if ex.ID != id {
		t.Errorf("exit message ID did not unwrap correctly, got %x expected %x",
			ex.ID, id)
		t.FailNow()
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
		on := Skins{}.
			Forward(&ap).
			Assemble()
		onb := Encode(on)
		c := slice.NewCursor()
		var onf types.Onion
		if onf, e = Peel(onb, c); check(e) {
			t.FailNow()
		}
		var cf *forward.Layer
		if cf, ok = onf.(*forward.Layer); !ok {
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
	on := Skins{}.
		Crypt(pub1, nil, prv2, n1, 0).
		Confirmation(n).
		Assemble()
	onb := Encode(on)
	c := slice.NewCursor()
	var onos, onc types.Onion
	if onos, e = Peel(onb, c); check(e) {
		t.Error(e)
		t.FailNow()
	}
	os := &crypt.Layer{}
	var ok bool
	if os, ok = onos.(*crypt.Layer); !ok {
		t.Error("did not unwrap expected type")
		t.FailNow()
	}
	os.Decrypt(prv1, onb, c)
	// unwrap the confirmation
	if onc, e = Peel(onb, c); check(e) {
		t.Error(e)
		t.FailNow()
	}
	oc := &confirm.Layer{}
	if oc, ok = onc.(*confirm.Layer); !ok {
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
		on := Skins{}.
			Reverse(&ap).
			Assemble()
		onb := Encode(on)
		c := slice.NewCursor()
		var onf types.Onion
		if onf, e = Peel(onb, c); check(e) {
			t.FailNow()
		}
		var cf *reverse.Layer
		if cf, ok = onf.(*reverse.Layer); !ok {
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
	var id nonce.ID
	var hash sha256.Hash
	if msg, hash, e = tests.GenMessage(10000, ""); check(e) {
		t.Error(e)
		t.FailNow()
	}
	on := Skins{}.
		Response(id, msg, 0).
		Assemble()
	onb := Encode(on)
	c := slice.NewCursor()
	var onex types.Onion
	if onex, e = Peel(onb, c); check(e) {
		t.FailNow()
	}
	ex := &response.Layer{}
	var ok bool
	if ex, ok = onex.(*response.Layer); !ok {
		t.Error("did not unwrap expected type")
		t.FailNow()
	}
	plH := sha256.Single(ex.Bytes)
	if plH != hash {
		t.Errorf("exit message did not unwrap correctly")
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
