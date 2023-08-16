package Time

import (
	"testing"
	"time"
)

func TestTime(t *testing.T) {
	t1 := New().Put(time.Now())
	t2 := New()
	t2.Decode(t1.Encode())
	if t1.Get() != t2.Get() {
		t.Fail()
	}
}
