package node

import (
	"fmt"
	"testing"
)

func TestNodes_Add(t *testing.T) {
	n := NewNodes()
	const nNodes = 10000
	for i := 0; i < nNodes; i++ {
		var nn *Node
		nn = New(nil)
		n = n.Add(nn)
	}
	if n.Len() != nNodes {
		t.Error("did not create expected number of nodes")
	}
}

func TestNodes_DeleteByID(t *testing.T) {
	n := NewNodes()
	const nNodes = 10000
	var e error
	for i := 0; i < nNodes; i++ {
		var nn *Node
		nn = New(nil)
		n.Add(nn)
	}
	for i := range n {
		if n, e = n.DeleteByID(n[nNodes-i-1].ID); check(e) {
			t.Error(e)
		}
	}
}

func TestNodes_DeleteByIP(t *testing.T) {
	n := NewNodes()
	const nNodes = 10000
	var e error
	for i := 0; i < nNodes; i++ {
		var nn *Node
		nn = New(nil)
		n.Add(nn)
	}
	for i := range n {
		if n, e = n.DeleteByIP(n[nNodes-i-1].IP); check(e) {
			t.Error(e)
		}
	}
}

func TestNodes_FindByID(t *testing.T) {
	n := NewNodes()
	const nNodes = 10000
	for i := 0; i < nNodes; i++ {
		var nn *Node
		nn = New(nil)
		n.Add(nn)
	}
	for i := range n {
		if n.FindByID(n[nNodes-i-1].ID) == nil {
			t.Error(fmt.Errorf("id %v not found",
				n[nNodes-i-1].ID))
		}
	}
}

func TestNodes_FindByIP(t *testing.T) {
	n := NewNodes()
	const nNodes = 10000
	for i := 0; i < nNodes; i++ {
		var nn *Node
		nn = New(nil)
		n.Add(nn)
	}
	for i := range n {
		if n.FindByIP(n[nNodes-i-1].IP) == nil {
			t.Error(fmt.Errorf("id %v not found",
				n[nNodes-i-1].IP))
		}
	}
}
