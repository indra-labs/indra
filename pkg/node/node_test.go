package node

import (
	"fmt"
	"testing"
)

func TestNodes_Add(t *testing.T) {
	n := NewNodes()
	const nNodes = 10000
	var e error
	for i := 0; i < nNodes; i++ {
		var nn *Node
		if nn, e = New(nil); check(e) {
			t.Error(e)
		}
		n.Add(nn)
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
		if nn, e = New(nil); check(e) {
			t.Error(e)
		}
		n.Add(nn)
	}
	for i := range n.S {
		if e = n.DeleteByID(n.S[nNodes-i-1].ID); check(e) {
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
		if nn, e = New(nil); check(e) {
			t.Error(e)
		}
		n.Add(nn)
	}
	for i := range n.S {
		if e = n.DeleteByIP(n.S[nNodes-i-1].IP); check(e) {
			t.Error(e)
		}
	}
}

func TestNodes_FindByID(t *testing.T) {
	n := NewNodes()
	const nNodes = 10000
	var e error
	for i := 0; i < nNodes; i++ {
		var nn *Node
		if nn, e = New(nil); check(e) {
			t.Error(e)
		}
		n.Add(nn)
	}
	for i := range n.S {
		if n.FindByID(n.S[nNodes-i-1].ID) == nil {
			t.Error(fmt.Errorf("id %v not found",
				n.S[nNodes-i-1].ID))
		}
	}
}

func TestNodes_FindByIP(t *testing.T) {
	n := NewNodes()
	const nNodes = 10000
	var e error
	for i := 0; i < nNodes; i++ {
		var nn *Node
		if nn, e = New(nil); check(e) {
			t.Error(e)
		}
		n.Add(nn)
	}
	for i := range n.S {
		if n.FindByIP(n.S[nNodes-i-1].IP) == nil {
			t.Error(fmt.Errorf("id %v not found",
				n.S[nNodes-i-1].IP))
		}
	}
}
