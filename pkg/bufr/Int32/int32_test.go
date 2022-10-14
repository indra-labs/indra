package Int32

import (
	"strconv"
	"testing"
)

func TestInt32(t *testing.T) {
	exStr := "-456107345"
	ex, err := strconv.ParseInt(exStr, 10, 32)
	example := int32(ex)
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
