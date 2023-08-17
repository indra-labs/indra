package i32

import (
	"encoding/binary"
	"encoding/hex"
	"testing"
)

func TestI(t *testing.T) {
	by, err := hex.DecodeString("deadbeef")
	if err != nil {
		panic(err)
	}
	bits := binary.BigEndian.Uint32(by)
	bt := New()
	bt.Put(int32(bits))
	bt2 := New()
	bt2.Decode(bt.Encode())
	if bt.Get() != bt2.Get() {
		t.Fail()
	}
}
