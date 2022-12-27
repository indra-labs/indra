package wire

import (
	"reflect"
	"testing"

	"github.com/Indra-Labs/indra/pkg/key/pub"
	"github.com/Indra-Labs/indra/pkg/key/signer"
	"github.com/Indra-Labs/indra/pkg/node"
	"github.com/Indra-Labs/indra/pkg/nonce"
	"github.com/Indra-Labs/indra/pkg/slice"
	"github.com/Indra-Labs/indra/pkg/types"
	"github.com/Indra-Labs/indra/pkg/wire/cipher"
	"github.com/Indra-Labs/indra/pkg/wire/confirmation"
	"github.com/Indra-Labs/indra/pkg/wire/forward"
	"github.com/Indra-Labs/indra/pkg/wire/layer"
	log2 "github.com/cybriq/proc/pkg/log"
)

func PeelForward(t *testing.T, b slice.Bytes,
	c *slice.Cursor) (fwd *forward.OnionSkin) {

	var ok bool
	var on types.Onion
	var e error
	if on, e = PeelOnion(b, c); check(e) {
		t.Error(e)
		t.FailNow()
	}
	if fwd, ok = on.(*forward.OnionSkin); !ok {
		t.Error("did not unwrap expected type", reflect.TypeOf(fwd))
		t.FailNow()
	}
	return
}

func PeelOnionSkin(t *testing.T, b slice.Bytes,
	c *slice.Cursor) (l *layer.OnionSkin) {

	var ok bool
	var on types.Onion
	var e error
	if on, e = PeelOnion(b, c); check(e) {
		t.Error(e)
		t.FailNow()
	}
	if l, ok = on.(*layer.OnionSkin); !ok {
		t.Error("did not unwrap expected type", reflect.TypeOf(l))
		t.FailNow()
	}
	return
}

func PeelConfirmation(t *testing.T, b slice.Bytes,
	c *slice.Cursor) (cn *confirmation.OnionSkin) {

	var ok bool
	var e error
	var on types.Onion
	if on, e = PeelOnion(b, c); check(e) {
		t.Error(e)
		t.FailNow()
	}
	if cn, ok = on.(*confirmation.OnionSkin); !ok {
		t.Error("did not unwrap expected type", reflect.TypeOf(on))
		t.FailNow()
	}
	return
}

func TestPing(t *testing.T) {
	log2.CodeLoc = true
	_, ks, e := signer.New()
	if check(e) {
		t.Error(e)
		t.FailNow()
	}
	var hop [3]*node.Node
	for i := range hop {
		prv1, prv2 := GetTwoPrvKeys(t)
		pub1, pub2 := pub.Derive(prv1), pub.Derive(prv2)
		var n nonce.ID
		hop[i], n = node.New(slice.GenerateRandomAddrPortIPv4(),
			pub1, pub2, prv1, prv2, nil)
		_ = n
	}
	cprv1, cprv2 := GetTwoPrvKeys(t)
	cpub1, cpub2 := pub.Derive(cprv1), pub.Derive(cprv2)
	var n nonce.ID
	var client *node.Node
	client, n = node.New(slice.GenerateRandomAddrPortIPv4(),
		cpub1, cpub2, cprv1, cprv2, nil)

	on := Ping(n, client, hop, ks)
	b := EncodeOnion(on)
	c := slice.NewCursor()

	// Forward(hop[0].AddrPort).
	f0 := PeelForward(t, b, c)
	if hop[0].AddrPort.String() != f0.AddrPort.String() {
		t.Errorf("failed to unwrap expected: '%s', got '%s'",
			hop[0].AddrPort.String(), f0.AddrPort.String())
		t.FailNow()
	}

	// OnionSkin(address.FromPubKey(hop[0].HeaderKey), set.Next()).
	PeelOnionSkin(t, b, c).Decrypt(hop[0].HeaderPriv, b, c)

	// Forward(hop[1].AddrPort).
	f1 := PeelForward(t, b, c)
	if hop[1].AddrPort.String() != f1.AddrPort.String() {
		t.Errorf("failed to unwrap expected: '%s', got '%s'",
			hop[1].AddrPort.String(), f1.AddrPort.String())
		t.FailNow()
	}

	// OnionSkin(address.FromPubKey(hop[1].HeaderKey), set.Next()).
	PeelOnionSkin(t, b, c).Decrypt(hop[1].HeaderPriv, b, c)

	// Forward(hop[2].AddrPort).
	f2 := PeelForward(t, b, c)
	if hop[2].AddrPort.String() != f2.AddrPort.String() {
		t.Errorf("failed to unwrap expected: '%s', got '%s'",
			hop[2].AddrPort.String(), f2.AddrPort.String())
		t.FailNow()
	}

	// OnionSkin(address.FromPubKey(hop[2].HeaderKey), set.Next()).
	PeelOnionSkin(t, b, c).Decrypt(hop[2].HeaderPriv, b, c)

	// Forward(client.AddrPort).
	f3 := PeelForward(t, b, c)
	if client.AddrPort.String() != f3.AddrPort.String() {
		t.Errorf("failed to unwrap expected: '%s', got '%s'",
			client.AddrPort.String(), f3.AddrPort.String())
		t.FailNow()
	}

	// OnionSkin(address.FromPubKey(client.HeaderKey), set.Next()).
	PeelOnionSkin(t, b, c).Decrypt(client.HeaderPriv, b, c)

	// Confirmation(id).
	co := PeelConfirmation(t, b, c)
	if co.ID != n {
		t.Error("did not unwrap expected confirmation nonce")
		t.FailNow()

	}
}

func TestSendKeys(t *testing.T) {
	log2.CodeLoc = true
	_, ks, e := signer.New()
	if check(e) {
		t.Error(e)
		t.FailNow()
	}
	var hop [5]*node.Node
	for i := range hop {
		prv1, prv2 := GetTwoPrvKeys(t)
		pub1, pub2 := pub.Derive(prv1), pub.Derive(prv2)
		hop[i], _ = node.New(slice.GenerateRandomAddrPortIPv4(),
			pub1, pub2, prv1, prv2, nil)
	}
	cprv1, cprv2 := GetTwoPrvKeys(t)
	cpub1, cpub2 := pub.Derive(cprv1), pub.Derive(cprv2)
	var n nonce.ID
	var client *node.Node
	client, n = node.New(slice.GenerateRandomAddrPortIPv4(),
		cpub1, cpub2, cprv1, cprv2, nil)
	ciprv1, ciprv2 := GetTwoPrvKeys(t)
	cipub1, cipub2 := pub.Derive(ciprv1), pub.Derive(ciprv2)

	on := SendKeys(n, cipub1, cipub2, client, hop, ks)
	b := EncodeOnion(on)
	c := slice.NewCursor()
	var ok bool

	// Forward(hop[0].AddrPort).
	f0 := PeelForward(t, b, c)
	if hop[0].AddrPort.String() != f0.AddrPort.String() {
		t.Errorf("failed to unwrap expected: '%s', got '%s'",
			hop[0].AddrPort.String(), f0.AddrPort.String())
		t.FailNow()
	}

	// OnionSkin(address.FromPubKey(hop[0].HeaderKey), set.Next()).
	PeelOnionSkin(t, b, c).Decrypt(hop[0].HeaderPriv, b, c)

	// Forward(hop[1].AddrPort).
	f1 := PeelForward(t, b, c)
	if hop[1].AddrPort.String() != f1.AddrPort.String() {
		t.Errorf("failed to unwrap expected: '%s', got '%s'",
			hop[1].AddrPort.String(), f1.AddrPort.String())
		t.FailNow()
	}

	// OnionSkin(address.FromPubKey(hop[1].HeaderKey), set.Next()).
	PeelOnionSkin(t, b, c).Decrypt(hop[1].HeaderPriv, b, c)

	// Forward(hop[2].AddrPort).
	f2 := PeelForward(t, b, c)
	if hop[2].AddrPort.String() != f2.AddrPort.String() {
		t.Errorf("failed to unwrap expected: '%s', got '%s'",
			hop[2].AddrPort.String(), f2.AddrPort.String())
		t.FailNow()
	}

	// OnionSkin(address.FromPubKey(hop[2].HeaderKey), set.Next()).
	PeelOnionSkin(t, b, c).Decrypt(hop[2].HeaderPriv, b, c)

	// Cipher(hdr, pld).
	var onc types.Onion
	if onc, e = PeelOnion(b, c); check(e) {
		t.Error(e)
		t.FailNow()
	}
	var ci *cipher.OnionSkin
	if ci, ok = onc.(*cipher.OnionSkin); !ok {
		t.Error("did not unwrap expected type", reflect.TypeOf(onc))
		t.FailNow()
	}
	if !ci.Header.Equals(cipub1) {
		t.Error("did not unwrap header key")
		t.FailNow()
	}
	if !ci.Payload.Equals(cipub2) {
		t.Error("did not unwrap payload key")
		t.FailNow()
	}

	// Forward(hop[3].AddrPort).
	f3 := PeelForward(t, b, c)
	if hop[3].AddrPort.String() != f3.AddrPort.String() {
		t.Errorf("failed to unwrap expected: '%s', got '%s'",
			hop[3].AddrPort.String(), f3.AddrPort.String())
		t.FailNow()
	}

	// OnionSkin(address.FromPubKey(hop[3].HeaderKey), set.Next()).
	PeelOnionSkin(t, b, c).Decrypt(hop[3].HeaderPriv, b, c)

	// Forward(hop[4].AddrPort).
	f4 := PeelForward(t, b, c)
	if hop[4].AddrPort.String() != f4.AddrPort.String() {
		t.Errorf("failed to unwrap expected: '%s', got '%s'",
			hop[3].AddrPort.String(), f4.AddrPort.String())
		t.FailNow()
	}

	// OnionSkin(address.FromPubKey(hop[4].HeaderKey), set.Next()).
	PeelOnionSkin(t, b, c).Decrypt(hop[4].HeaderPriv, b, c)

	// Forward(client.AddrPort).
	f5 := PeelForward(t, b, c)
	if client.AddrPort.String() != f5.AddrPort.String() {
		t.Errorf("failed to unwrap expected: '%s', got '%s'",
			client.AddrPort.String(), f5.AddrPort.String())
		t.FailNow()
	}

	// OnionSkin(address.FromPubKey(client.HeaderKey), set.Next()).
	PeelOnionSkin(t, b, c).Decrypt(client.HeaderPriv, b, c)

	// Confirmation(id).
	co := PeelConfirmation(t, b, c)
	if co.ID != n {
		t.Error("did not unwrap expected confirmation nonce")
		t.FailNow()

	}

}
