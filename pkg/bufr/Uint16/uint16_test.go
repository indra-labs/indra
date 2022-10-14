package Uint16

import (
	"strconv"
	"testing"
)

func TestUint16(t *testing.T) {
	exStr := "11047"
	ex, err := strconv.ParseInt(exStr, 10, 16)
	example := uint16(ex)
	if err != nil {
		t.Fail()
	}
	u := New()
	u.Put(example)
	port2 := New()
	port2.Decode(u.Encode())
	if port2.Get() != u.Get() {
		t.Fail()
	}
	if u.String() != exStr {
		t.Log(u.String(), exStr)
		t.Fail()
	}
}
