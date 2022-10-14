package Uint64

import (
	"strconv"
	"testing"
)

func TestUint64(t *testing.T) {
	exStr := "1104734534534234"
	example, err := strconv.ParseUint(exStr, 10, 64)
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
