package Uint32

import (
	"strconv"
	"testing"
)

func TestUint32(t *testing.T) {
	exStr := "1107343455"
	ex, err := strconv.ParseUint(exStr, 10, 32)
	example := uint32(ex)
	if err != nil {
		t.Fail()
	}
	u := New()
	u.Put(example)
	u2 := New()
	u2.Decode(u.Encode())
	if u2.Get() != u.Get() {
		t.Fail()
	}
	if u.String() != exStr {
		t.Log(u.String(), exStr)
		t.Fail()
	}
}
