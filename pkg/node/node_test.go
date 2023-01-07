package node

import (
	"fmt"
	"net/netip"
	"testing"

	"github.com/indra-labs/indra/pkg/testutils"
	"github.com/indra-labs/indra/pkg/transport"
)

var testAddrPort, _ = netip.ParseAddrPort("1.1.1.1:20000")

func TestNodes_Add(t *testing.T) {
	n := NewNodes()
	pubKey, prvKey, e := testutils.GenerateTestKeyPair()
	if check(e) {
		t.Error(e)
		t.FailNow()
	}
	const nNodes = 10000
	for i := 0; i < nNodes; i++ {
		var nn *Node
		nn, _ = New(&testAddrPort, prvKey, pubKey, transport.NewSim(0))
		n = n.Add(nn)
	}
	if n.Len() != nNodes {
		t.Error("did not create expected number of nodes")
	}
}

func TestNodes_DeleteByID(t *testing.T) {
	n := NewNodes()
	const nNodes = 10000
	pubKey, prvKey, e := testutils.GenerateTestKeyPair()
	if check(e) {
		t.Error(e)
		t.FailNow()
	}
	for i := 0; i < nNodes; i++ {
		var nn *Node
		nn, _ = New(&testAddrPort, prvKey, pubKey, transport.NewSim(0))
		n.Add(nn)
	}
	for i := range n {
		if n, e = n.DeleteByID(n[nNodes-i-1].ID); check(e) {
			t.Error(e)
		}
	}
}

func TestNodes_DeleteByAddrPort(t *testing.T) {
	n := NewNodes()
	const nNodes = 10000
	pubKey, prvKey, e := testutils.GenerateTestKeyPair()
	if check(e) {
		t.Error(e)
		t.FailNow()
	}
	for i := 0; i < nNodes; i++ {
		var nn *Node
		nn, _ = New(&testAddrPort, prvKey, pubKey, transport.NewSim(0))
		n.Add(nn)
	}
	for i := range n {
		if n, e = n.DeleteByAddrPort(n[nNodes-i-1].AddrPort); check(e) {
			t.Error(e)
		}
	}
}

func TestNodes_FindByID(t *testing.T) {
	n := NewNodes()
	const nNodes = 10000
	pubKey, prvKey, e := testutils.GenerateTestKeyPair()
	if check(e) {
		t.Error(e)
		t.FailNow()
	}
	for i := 0; i < nNodes; i++ {
		var nn *Node
		nn, _ = New(&testAddrPort, prvKey, pubKey, transport.NewSim(0))
		n.Add(nn)
	}
	for i := range n {
		if n.FindByID(n[nNodes-i-1].ID) == nil {
			t.Error(fmt.Errorf("id %v not found",
				n[nNodes-i-1].ID))
		}
	}
}

func TestNodes_FindByAddrPort(t *testing.T) {
	n := NewNodes()
	const nNodes = 10000
	pubKey, prvKey, e := testutils.GenerateTestKeyPair()
	if check(e) {
		t.Error(e)
		t.FailNow()
	}
	for i := 0; i < nNodes; i++ {
		var nn *Node
		nn, _ = New(&testAddrPort, prvKey, pubKey, transport.NewSim(0))
		n.Add(nn)
	}
	for i := range n {
		if n.FindByAddrPort(n[nNodes-i-1].AddrPort) == nil {
			t.Error(fmt.Errorf("id %v not found",
				n[nNodes-i-1].AddrPort))
		}
	}
}
