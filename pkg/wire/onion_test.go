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

	var ok bool
	var on0 types.Onion
	if on0, e = PeelOnion(b, c); check(e) {
		t.Error(e)
		t.FailNow()
	}
	var f0 *forward.OnionSkin
	if f0, ok = on0.(*forward.OnionSkin); !ok {
		t.Error("did not unwrap expected type", reflect.TypeOf(f0))
		t.FailNow()
	}
	if hop[0].AddrPort.String() != f0.AddrPort.String() {
		t.Errorf("failed to unwrap expected: '%s', got '%s'",
			hop[0].AddrPort.String(), f0.AddrPort.String())
		t.FailNow()
	}

	var on1 types.Onion
	if on1, e = PeelOnion(b, c); check(e) {
		t.Error(e)
		t.FailNow()
	}
	var l0 *layer.OnionSkin
	if l0, ok = on1.(*layer.OnionSkin); !ok {
		t.Error("did not unwrap expected type", reflect.TypeOf(l0))
		t.FailNow()
	}
	l0.Decrypt(hop[0].HeaderPriv, b, c)
	var on2 types.Onion
	if on2, e = PeelOnion(b, c); check(e) {
		t.Error(e)
		t.FailNow()
	}
	var f1 *forward.OnionSkin
	if f1, ok = on2.(*forward.OnionSkin); !ok {
		t.Error("did not unwrap expected type", reflect.TypeOf(on2))
		t.FailNow()
	}
	if hop[1].AddrPort.String() != f1.AddrPort.String() {
		t.Errorf("failed to unwrap expected: '%s', got '%s'",
			hop[1].AddrPort.String(), f1.AddrPort.String())
		t.FailNow()
	}

	var on3 types.Onion
	if on3, e = PeelOnion(b, c); check(e) {
		t.Error(e)
		t.FailNow()
	}
	var l1 *layer.OnionSkin
	if l1, ok = on3.(*layer.OnionSkin); !ok {
		t.Error("did not unwrap expected type", reflect.TypeOf(l1))
		t.FailNow()
	}
	l1.Decrypt(hop[1].HeaderPriv, b, c)
	var on4 types.Onion
	if on4, e = PeelOnion(b, c); check(e) {
		t.Error(e)
		t.FailNow()
	}
	var f2 *forward.OnionSkin
	if f2, ok = on4.(*forward.OnionSkin); !ok {
		t.Error("did not unwrap expected type", reflect.TypeOf(on2))
		t.FailNow()
	}
	if hop[2].AddrPort.String() != f2.AddrPort.String() {
		t.Errorf("failed to unwrap expected: '%s', got '%s'",
			hop[2].AddrPort.String(), f2.AddrPort.String())
		t.FailNow()
	}

	var on5 types.Onion
	if on5, e = PeelOnion(b, c); check(e) {
		t.Error(e)
		t.FailNow()
	}
	var l2 *layer.OnionSkin
	if l2, ok = on5.(*layer.OnionSkin); !ok {
		t.Error("did not unwrap expected type", reflect.TypeOf(l1))
		t.FailNow()
	}
	l2.Decrypt(hop[2].HeaderPriv, b, c)
	var on6 types.Onion
	if on6, e = PeelOnion(b, c); check(e) {
		t.Error(e)
		t.FailNow()
	}
	var f3 *forward.OnionSkin
	if f3, ok = on6.(*forward.OnionSkin); !ok {
		t.Error("did not unwrap expected type", reflect.TypeOf(on2))
		t.FailNow()
	}
	if client.AddrPort.String() != f3.AddrPort.String() {
		t.Errorf("failed to unwrap expected: '%s', got '%s'",
			client.AddrPort.String(), f3.AddrPort.String())
		t.FailNow()
	}

	var on7 types.Onion
	if on7, e = PeelOnion(b, c); check(e) {
		t.Error(e)
		t.FailNow()
	}
	var l3 *layer.OnionSkin
	if l3, ok = on7.(*layer.OnionSkin); !ok {
		t.Error("did not unwrap expected type", reflect.TypeOf(l1))
		t.FailNow()
	}
	l3.Decrypt(client.HeaderPriv, b, c)
	var on8 types.Onion
	if on8, e = PeelOnion(b, c); check(e) {
		t.Error(e)
		t.FailNow()
	}
	var co *confirmation.OnionSkin
	if co, ok = on8.(*confirmation.OnionSkin); !ok {
		t.Error("did not unwrap expected type", reflect.TypeOf(on8))
		t.FailNow()
	}

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
	var on0 types.Onion
	if on0, e = PeelOnion(b, c); check(e) {
		t.Error(e)
		t.FailNow()
	}
	var f0 *forward.OnionSkin
	if f0, ok = on0.(*forward.OnionSkin); !ok {
		t.Error("did not unwrap expected type", reflect.TypeOf(on0))
		t.FailNow()
	}
	if hop[0].AddrPort.String() != f0.AddrPort.String() {
		t.Errorf("failed to unwrap expected: '%s', got '%s'",
			hop[0].AddrPort.String(), f0.AddrPort.String())
		t.FailNow()
	}

	// OnionSkin(address.FromPubKey(hop[0].HeaderKey), set.Next()).
	var on1 types.Onion
	if on1, e = PeelOnion(b, c); check(e) {
		t.Error(e)
		t.FailNow()
	}
	var l0 *layer.OnionSkin
	if l0, ok = on1.(*layer.OnionSkin); !ok {
		t.Error("did not unwrap expected type", reflect.TypeOf(l0))
		t.FailNow()
	}
	l0.Decrypt(hop[0].HeaderPriv, b, c)

	// Forward(hop[1].AddrPort).
	var on2 types.Onion
	if on2, e = PeelOnion(b, c); check(e) {
		t.Error(e)
		t.FailNow()
	}
	var f1 *forward.OnionSkin
	if f1, ok = on2.(*forward.OnionSkin); !ok {
		t.Error("did not unwrap expected type", reflect.TypeOf(on2))
		t.FailNow()
	}
	if hop[1].AddrPort.String() != f1.AddrPort.String() {
		t.Errorf("failed to unwrap expected: '%s', got '%s'",
			hop[1].AddrPort.String(), f1.AddrPort.String())
		t.FailNow()
	}

	// OnionSkin(address.FromPubKey(hop[1].HeaderKey), set.Next()).
	var on3 types.Onion
	if on3, e = PeelOnion(b, c); check(e) {
		t.Error(e)
		t.FailNow()
	}
	var l1 *layer.OnionSkin
	if l1, ok = on3.(*layer.OnionSkin); !ok {
		t.Error("did not unwrap expected type", reflect.TypeOf(l1))
		t.FailNow()
	}
	l1.Decrypt(hop[1].HeaderPriv, b, c)

	// Forward(hop[2].AddrPort).
	var on4 types.Onion
	if on4, e = PeelOnion(b, c); check(e) {
		t.Error(e)
		t.FailNow()
	}
	var f2 *forward.OnionSkin
	if f2, ok = on4.(*forward.OnionSkin); !ok {
		t.Error("did not unwrap expected type", reflect.TypeOf(on4))
		t.FailNow()
	}
	if hop[2].AddrPort.String() != f2.AddrPort.String() {
		t.Errorf("failed to unwrap expected: '%s', got '%s'",
			hop[2].AddrPort.String(), f2.AddrPort.String())
		t.FailNow()
	}

	// OnionSkin(address.FromPubKey(hop[2].HeaderKey), set.Next()).
	var on5 types.Onion
	if on5, e = PeelOnion(b, c); check(e) {
		t.Error(e)
		t.FailNow()
	}
	var l2 *layer.OnionSkin
	if l2, ok = on5.(*layer.OnionSkin); !ok {
		t.Error("did not unwrap expected type", reflect.TypeOf(l1))
		t.FailNow()
	}
	l2.Decrypt(hop[2].HeaderPriv, b, c)

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
	var on6 types.Onion
	if on6, e = PeelOnion(b, c); check(e) {
		t.Error(e)
		t.FailNow()
	}
	var f3 *forward.OnionSkin
	if f3, ok = on6.(*forward.OnionSkin); !ok {
		t.Error("did not unwrap expected type", reflect.TypeOf(on6))
		t.FailNow()
	}
	if hop[3].AddrPort.String() != f3.AddrPort.String() {
		t.Errorf("failed to unwrap expected: '%s', got '%s'",
			hop[3].AddrPort.String(), f3.AddrPort.String())
		t.FailNow()
	}

	// OnionSkin(address.FromPubKey(hop[3].HeaderKey), set.Next()).
	var on7 types.Onion
	if on7, e = PeelOnion(b, c); check(e) {
		t.Error(e)
		t.FailNow()
	}
	var l3 *layer.OnionSkin
	if l3, ok = on7.(*layer.OnionSkin); !ok {
		t.Error("did not unwrap expected type", reflect.TypeOf(l3))
		t.FailNow()
	}
	l3.Decrypt(hop[3].HeaderPriv, b, c)

	// Forward(hop[4].AddrPort).
	var on8 types.Onion
	if on8, e = PeelOnion(b, c); check(e) {
		t.Error(e)
		t.FailNow()
	}
	var f4 *forward.OnionSkin
	if f4, ok = on8.(*forward.OnionSkin); !ok {
		t.Error("did not unwrap expected type", reflect.TypeOf(on8))
		t.FailNow()
	}
	if hop[4].AddrPort.String() != f4.AddrPort.String() {
		t.Errorf("failed to unwrap expected: '%s', got '%s'",
			hop[3].AddrPort.String(), f4.AddrPort.String())
		t.FailNow()
	}

	// OnionSkin(address.FromPubKey(hop[4].HeaderKey), set.Next()).
	var on9 types.Onion
	if on9, e = PeelOnion(b, c); check(e) {
		t.Error(e)
		t.FailNow()
	}
	var l4 *layer.OnionSkin
	if l4, ok = on9.(*layer.OnionSkin); !ok {
		t.Error("did not unwrap expected type", reflect.TypeOf(l4))
		t.FailNow()
	}
	l4.Decrypt(hop[4].HeaderPriv, b, c)

	// Forward(client.AddrPort).
	var on10 types.Onion
	if on10, e = PeelOnion(b, c); check(e) {
		t.Error(e)
		t.FailNow()
	}
	var f5 *forward.OnionSkin
	if f5, ok = on10.(*forward.OnionSkin); !ok {
		t.Error("did not unwrap expected type", reflect.TypeOf(on10))
		t.FailNow()
	}
	if client.AddrPort.String() != f5.AddrPort.String() {
		t.Errorf("failed to unwrap expected: '%s', got '%s'",
			client.AddrPort.String(), f5.AddrPort.String())
		t.FailNow()
	}

	// OnionSkin(address.FromPubKey(client.HeaderKey), set.Next()).
	var on11 types.Onion
	if on11, e = PeelOnion(b, c); check(e) {
		t.Error(e)
		t.FailNow()
	}
	var l5 *layer.OnionSkin
	if l5, ok = on11.(*layer.OnionSkin); !ok {
		t.Error("did not unwrap expected type", reflect.TypeOf(l5))
		t.FailNow()
	}
	l5.Decrypt(client.HeaderPriv, b, c)

	// Confirmation(id).
	var on12 types.Onion
	if on12, e = PeelOnion(b, c); check(e) {
		t.Error(e)
		t.FailNow()
	}
	var co *confirmation.OnionSkin
	if co, ok = on12.(*confirmation.OnionSkin); !ok {
		t.Error("did not unwrap expected type", reflect.TypeOf(on12))
		t.FailNow()
	}

	if co.ID != n {
		t.Error("did not unwrap expected confirmation nonce")
		t.FailNow()

	}

}
