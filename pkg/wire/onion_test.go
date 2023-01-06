package wire

import (
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/indra-labs/indra/pkg/key/pub"
	"github.com/indra-labs/indra/pkg/key/signer"
	log2 "github.com/indra-labs/indra/pkg/log"
	"github.com/indra-labs/indra/pkg/node"
	"github.com/indra-labs/indra/pkg/nonce"
	"github.com/indra-labs/indra/pkg/session"
	"github.com/indra-labs/indra/pkg/sha256"
	"github.com/indra-labs/indra/pkg/slice"
	"github.com/indra-labs/indra/pkg/testutils"
	"github.com/indra-labs/indra/pkg/types"
	"github.com/indra-labs/indra/pkg/wire/cipher"
	"github.com/indra-labs/indra/pkg/wire/confirm"
	"github.com/indra-labs/indra/pkg/wire/exit"
	"github.com/indra-labs/indra/pkg/wire/forward"
	"github.com/indra-labs/indra/pkg/wire/layer"
	"github.com/indra-labs/indra/pkg/wire/purchase"
	"github.com/indra-labs/indra/pkg/wire/reverse"
)

func PeelForward(t *testing.T, b slice.Bytes,
	c *slice.Cursor) (fwd *forward.OnionSkin) {

	var ok bool
	var on types.Onion
	var e error
	if on, e = PeelOnion(b, c); check(e) {
		t.Error(e)
	}
	if fwd, ok = on.(*forward.OnionSkin); !ok {
		t.Error("did not unwrap expected type", reflect.TypeOf(fwd))
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
	}
	if l, ok = on.(*layer.OnionSkin); !ok {
		t.Error("did not unwrap expected type", reflect.TypeOf(l))
	}
	return
}

func PeelConfirmation(t *testing.T, b slice.Bytes,
	c *slice.Cursor) (cn *confirm.OnionSkin) {

	var ok bool
	var e error
	var on types.Onion
	if on, e = PeelOnion(b, c); check(e) {
		t.Error(e)
	}
	if cn, ok = on.(*confirm.OnionSkin); !ok {
		t.Error("did not unwrap expected type", reflect.TypeOf(on))
	}
	return
}

func PeelPurchase(t *testing.T, b slice.Bytes,
	c *slice.Cursor) (pr *purchase.OnionSkin) {

	var ok bool
	var e error
	var on types.Onion
	if on, e = PeelOnion(b, c); check(e) {
		t.Error(e)
	}
	if pr, ok = on.(*purchase.OnionSkin); !ok {
		t.Error("did not unwrap expected type", reflect.TypeOf(on))
	}
	return
}

func PeelReverse(t *testing.T, b slice.Bytes,
	c *slice.Cursor) (rp *reverse.OnionSkin) {

	var ok bool
	var e error
	var on types.Onion
	if on, e = PeelOnion(b, c); check(e) {
		t.Error(e)
	}
	if rp, ok = on.(*reverse.OnionSkin); !ok {
		t.Error("did not unwrap expected type", reflect.TypeOf(on))
	}
	return
}

func PeelExit(t *testing.T, b slice.Bytes,
	c *slice.Cursor) (ex *exit.OnionSkin) {

	var ok bool
	var e error
	var on types.Onion
	if on, e = PeelOnion(b, c); check(e) {
		t.Error(e)
		t.FailNow()
	}
	if ex, ok = on.(*exit.OnionSkin); !ok {
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
		prv1, _ := GetTwoPrvKeys(t)
		pub1 := pub.Derive(prv1)
		var n nonce.ID
		hop[i], n = node.New(slice.GenerateRandomAddrPortIPv4(),
			pub1, prv1, nil)
		_ = n
	}
	cprv1, _ := GetTwoPrvKeys(t)
	cpub1 := pub.Derive(cprv1)
	var n nonce.ID
	var client *node.Node
	client, n = node.New(slice.GenerateRandomAddrPortIPv4(),
		cpub1, cprv1, nil)

	on := Ping(n, client, hop, ks)
	b := EncodeOnion(on.Assemble())
	c := slice.NewCursor()

	// Forward(hop[0].AddrPort).
	f0 := PeelForward(t, b, c)
	if hop[0].AddrPort.String() != f0.AddrPort.String() {
		t.Errorf("failed to unwrap; expected: '%s', got: '%s'",
			hop[0].AddrPort.String(), f0.AddrPort.String())
		t.FailNow()
	}

	// OnionSkin(address.FromPubKey(hop[0].HeaderPub), set.Next()).
	PeelOnionSkin(t, b, c).Decrypt(hop[0].HeaderPrv, b, c)

	// Forward(hop[1].AddrPort).
	f1 := PeelForward(t, b, c)
	if hop[1].AddrPort.String() != f1.AddrPort.String() {
		t.Errorf("failed to unwrap; expected: '%s', got: '%s'",
			hop[1].AddrPort.String(), f1.AddrPort.String())
		t.FailNow()
	}

	// OnionSkin(address.FromPubKey(hop[1].HeaderPub), set.Next()).
	PeelOnionSkin(t, b, c).Decrypt(hop[1].HeaderPrv, b, c)

	// Forward(hop[2].AddrPort).
	f2 := PeelForward(t, b, c)
	if hop[2].AddrPort.String() != f2.AddrPort.String() {
		t.Errorf("failed to unwrap; expected: '%s', got: '%s'",
			hop[2].AddrPort.String(), f2.AddrPort.String())
		t.FailNow()
	}

	// OnionSkin(address.FromPubKey(hop[2].HeaderPub), set.Next()).
	PeelOnionSkin(t, b, c).Decrypt(hop[2].HeaderPrv, b, c)

	// Forward(client.AddrPort).
	f3 := PeelForward(t, b, c)
	if client.AddrPort.String() != f3.AddrPort.String() {
		t.Errorf("failed to unwrap; expected: '%s', got: '%s'",
			client.AddrPort.String(), f3.AddrPort.String())
		t.FailNow()
	}

	// OnionSkin(address.FromPubKey(client.HeaderPub), set.Next()).
	PeelOnionSkin(t, b, c).Decrypt(client.HeaderPrv, b, c)

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
		prv1, _ := GetTwoPrvKeys(t)
		pub1 := pub.Derive(prv1)
		hop[i], _ = node.New(slice.GenerateRandomAddrPortIPv4(),
			pub1, prv1, nil)
	}
	cprv1, _ := GetTwoPrvKeys(t)
	cpub1 := pub.Derive(cprv1)
	var n nonce.ID
	var client *node.Node
	client, n = node.New(slice.GenerateRandomAddrPortIPv4(),
		cpub1, cprv1, nil)
	ciprv1, ciprv2 := GetTwoPrvKeys(t)
	cipub1, cipub2 := pub.Derive(ciprv1), pub.Derive(ciprv2)

	on := SendKeys(n, cipub1, cipub2, client, hop, ks)
	b := EncodeOnion(on.Assemble())
	c := slice.NewCursor()
	var ok bool

	// Forward(hop[0].AddrPort).
	f0 := PeelForward(t, b, c)
	if hop[0].AddrPort.String() != f0.AddrPort.String() {
		t.Errorf("failed to unwrap expected: '%s', got '%s'",
			hop[0].AddrPort.String(), f0.AddrPort.String())
		t.FailNow()
	}

	// OnionSkin(address.FromPubKey(hop[0].HeaderPub), set.Next()).
	PeelOnionSkin(t, b, c).Decrypt(hop[0].HeaderPrv, b, c)

	// Forward(hop[1].AddrPort).
	f1 := PeelForward(t, b, c)
	if hop[1].AddrPort.String() != f1.AddrPort.String() {
		t.Errorf("failed to unwrap expected: '%s', got '%s'",
			hop[1].AddrPort.String(), f1.AddrPort.String())
		t.FailNow()
	}

	// OnionSkin(address.FromPubKey(hop[1].HeaderPub), set.Next()).
	PeelOnionSkin(t, b, c).Decrypt(hop[1].HeaderPrv, b, c)

	// Forward(hop[2].AddrPort).
	f2 := PeelForward(t, b, c)
	if hop[2].AddrPort.String() != f2.AddrPort.String() {
		t.Errorf("failed to unwrap expected: '%s', got '%s'",
			hop[2].AddrPort.String(), f2.AddrPort.String())
		t.FailNow()
	}

	// OnionSkin(address.FromPubKey(hop[2].HeaderPub), set.Next()).
	PeelOnionSkin(t, b, c).Decrypt(hop[2].HeaderPrv, b, c)

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

	// OnionSkin(address.FromPubKey(hop[3].HeaderPub), set.Next()).
	PeelOnionSkin(t, b, c).Decrypt(hop[3].HeaderPrv, b, c)

	// Forward(hop[4].AddrPort).
	f4 := PeelForward(t, b, c)
	if hop[4].AddrPort.String() != f4.AddrPort.String() {
		t.Errorf("failed to unwrap expected: '%s', got '%s'",
			hop[3].AddrPort.String(), f4.AddrPort.String())
		t.FailNow()
	}

	// OnionSkin(address.FromPubKey(hop[4].HeaderPub), set.Next()).
	PeelOnionSkin(t, b, c).Decrypt(hop[4].HeaderPrv, b, c)

	// Forward(client.AddrPort).
	f5 := PeelForward(t, b, c)
	if client.AddrPort.String() != f5.AddrPort.String() {
		t.Errorf("failed to unwrap expected: '%s', got '%s'",
			client.AddrPort.String(), f5.AddrPort.String())
		t.FailNow()
	}

	// OnionSkin(address.FromPubKey(client.HeaderPub), set.Next()).
	PeelOnionSkin(t, b, c).Decrypt(client.HeaderPrv, b, c)

	// Confirmation(id).
	co := PeelConfirmation(t, b, c)
	if co.ID != n {
		t.Error("did not unwrap expected confirmation nonce")
		t.FailNow()

	}

}

func TestSendPurchase(t *testing.T) {
	log2.CodeLoc = true
	log2.SetLogLevel(log2.Trace)
	_, ks, e := signer.New()
	if check(e) {
		t.Error(e)
		t.FailNow()
	}
	var hop [5]*node.Node
	for i := range hop {
		prv1, _ := GetTwoPrvKeys(t)
		pub1 := pub.Derive(prv1)
		hop[i], _ = node.New(slice.GenerateRandomAddrPortIPv4(),
			pub1, prv1, nil)
	}
	cprv1, _ := GetTwoPrvKeys(t)
	cpub1 := pub.Derive(cprv1)
	var client *node.Node
	client, _ = node.New(slice.GenerateRandomAddrPortIPv4(),
		cpub1, cprv1, nil)
	var sess [3]*session.Session
	for i := range sess {
		sess[i] = session.NewSession(nonce.NewID(), 203230230,
			time.Hour, ks)
	}
	nBytes := rand.Uint64()
	n := nonce.NewID()
	on := SendPurchase(n, nBytes, client, hop, sess, ks)
	b := EncodeOnion(on.Assemble())
	c := slice.NewCursor()

	// Forward(hop[0].AddrPort).
	f0 := PeelForward(t, b, c)
	if hop[0].AddrPort.String() != f0.AddrPort.String() {
		t.Errorf("failed to unwrap expected: '%s', got '%s'",
			hop[0].AddrPort.String(), f0.AddrPort.String())
		t.FailNow()
	}

	// OnionSkin(address.FromPubKey(hop[0].HeaderPub), set.Next()).
	PeelOnionSkin(t, b, c).Decrypt(hop[0].HeaderPrv, b, c)

	// Forward(hop[1].AddrPort).
	f1 := PeelForward(t, b, c)
	if hop[1].AddrPort.String() != f1.AddrPort.String() {
		t.Errorf("failed to unwrap expected: '%s', got '%s'",
			hop[0].AddrPort.String(), f1.AddrPort.String())
		t.FailNow()
	}

	// OnionSkin(address.FromPubKey(hop[1].HeaderPub), set.Next()).
	PeelOnionSkin(t, b, c).Decrypt(hop[1].HeaderPrv, b, c)

	// Forward(hop[2].AddrPort).
	f2 := PeelForward(t, b, c)
	if hop[2].AddrPort.String() != f2.AddrPort.String() {
		t.Errorf("failed to unwrap expected: '%s', got '%s'",
			hop[1].AddrPort.String(), f1.AddrPort.String())
		t.FailNow()
	}

	// OnionSkin(address.FromPubKey(hop[2].HeaderPub), set.Next()).
	PeelOnionSkin(t, b, c).Decrypt(hop[2].HeaderPrv, b, c)

	// Purchase(nBytes, prvs, pubs).
	pr := PeelPurchase(t, b, c)
	if pr.NBytes != nBytes {
		t.Errorf("failed to retrieve original purchase nBytes")
		t.FailNow()
	}

	// Reverse(hop[3].AddrPort).
	rp1 := PeelReverse(t, b, c)
	if rp1.AddrPort.String() != hop[3].AddrPort.String() {
		t.Errorf("failed to retrieve first reply hop")
		t.FailNow()
	}

	// OnionSkin(address.FromPubKey(hop[3].HeaderPub), replies[0]).
	PeelOnionSkin(t, b, c).Decrypt(sess[0].HeaderPrv, b, c)

	// Reverse(hop[4].AddrPort).
	rp2 := PeelReverse(t, b, c)
	if rp2.AddrPort.String() != hop[4].AddrPort.String() {
		t.Errorf("failed to retrieve second reply hop")
		t.FailNow()
	}

	// OnionSkin(address.FromPubKey(hop[4].HeaderPub), replies[1]).
	PeelOnionSkin(t, b, c).Decrypt(sess[1].HeaderPrv, b, c)

	// Reverse(client.AddrPort).
	rp3 := PeelReverse(t, b, c)
	if rp3.AddrPort.String() != client.AddrPort.String() {
		t.Errorf("failed to retrieve third reply hop")
		t.FailNow()
	}

	// OnionSkin(address.FromPubKey(client.HeaderPub), replies[2]).
	PeelOnionSkin(t, b, c).Decrypt(sess[2].HeaderPrv, b, c)

}

func TestSendExit(t *testing.T) {
	log2.CodeLoc = true
	_, ks, e := signer.New()
	if check(e) {
		t.Error(e)
		t.FailNow()
	}
	var hop [5]*node.Node
	for i := range hop {
		prv1, _ := GetTwoPrvKeys(t)
		pub1 := pub.Derive(prv1)
		hop[i], _ = node.New(slice.GenerateRandomAddrPortIPv4(),
			pub1, prv1, nil)
	}
	cprv1, _ := GetTwoPrvKeys(t)
	cpub1 := pub.Derive(cprv1)
	var client *node.Node
	client, _ = node.New(slice.GenerateRandomAddrPortIPv4(),
		cpub1, cprv1, nil)
	port := uint16(rand.Uint32())
	var message slice.Bytes
	var hash sha256.Hash
	message, hash, e = testutils.GenerateTestMessage(2502)
	var sess [3]*session.Session
	for i := range sess {
		sess[i] = session.NewSession(nonce.NewID(), 203230230,
			time.Hour, ks)
	}
	on := SendExit(message, port, client, hop, sess, ks)
	b := EncodeOnion(on.Assemble())
	c := slice.NewCursor()

	// Forward(hop[0].AddrPort).
	f0 := PeelForward(t, b, c)
	if hop[0].AddrPort.String() != f0.AddrPort.String() {
		t.Errorf("failed to unwrap expected: '%s', got '%s'",
			hop[0].AddrPort.String(), f0.AddrPort.String())
		t.FailNow()
	}

	// OnionSkin(address.FromPubKey(hop[0].HeaderPub), set.Next()).
	PeelOnionSkin(t, b, c).Decrypt(hop[0].HeaderPrv, b, c)

	// Forward(hop[1].AddrPort).
	f1 := PeelForward(t, b, c)
	if hop[1].AddrPort.String() != f1.AddrPort.String() {
		t.Errorf("failed to unwrap expected: '%s', got '%s'",
			hop[0].AddrPort.String(), f1.AddrPort.String())
		t.FailNow()
	}

	// OnionSkin(address.FromPubKey(hop[1].HeaderPub), set.Next()).
	PeelOnionSkin(t, b, c).Decrypt(hop[1].HeaderPrv, b, c)

	// Forward(hop[2].AddrPort).
	f2 := PeelForward(t, b, c)
	if hop[2].AddrPort.String() != f2.AddrPort.String() {
		t.Errorf("failed to unwrap expected: '%s', got '%s'",
			hop[1].AddrPort.String(), f1.AddrPort.String())
		t.FailNow()
	}

	// OnionSkin(address.FromPubKey(hop[2].HeaderPub), set.Next()).
	PeelOnionSkin(t, b, c).Decrypt(hop[2].HeaderPrv, b, c)

	// Exit(port, prvs, pubs, payload).
	pr := PeelExit(t, b, c)
	if pr.Port != port {
		t.Errorf("failed to retrieve original purchase nBytes")
		t.FailNow()
	}
	mh := sha256.Single(pr.Bytes)
	if mh != hash {
		t.Errorf("exit message not correctly decoded")
		t.FailNow()
	}

	// Reverse(hop[3].AddrPort).
	rp1 := PeelReverse(t, b, c)
	if rp1.AddrPort.String() != hop[3].AddrPort.String() {
		t.Errorf("failed to retrieve first reply hop")
		t.FailNow()
	}

	// OnionSkin(address.FromPubKey(hop[3].HeaderPub), replies[0]).
	PeelOnionSkin(t, b, c).Decrypt(sess[0].HeaderPrv, b, c)

	// Reverse(hop[4].AddrPort).
	rp2 := PeelReverse(t, b, c)
	if rp2.AddrPort.String() != hop[4].AddrPort.String() {
		t.Errorf("failed to retrieve second reply hop")
		t.FailNow()
	}

	// OnionSkin(address.FromPubKey(hop[4].HeaderPub), replies[1]).
	PeelOnionSkin(t, b, c).Decrypt(sess[1].HeaderPrv, b, c)

	// Reverse(client.AddrPort).
	rp3 := PeelReverse(t, b, c)
	if rp3.AddrPort.String() != client.AddrPort.String() {
		t.Errorf("failed to retrieve third reply hop")
		t.FailNow()
	}

	// OnionSkin(address.FromPubKey(client.HeaderPub), replies[2]).
	PeelOnionSkin(t, b, c).Decrypt(sess[2].HeaderPrv, b, c)

}
