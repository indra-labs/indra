package i32

import (
	log2 "git.indra-labs.org/dev/ind/pkg/proc/log"
	"git.indra-labs.org/dev/ind/pkg/util/ci"
	"testing"
)

var (
	log   = log2.GetLogger()
	fails = log.E.Chk
)

func TestS(t *testing.T) {
	ci.TraceIfNot()
	
	t1, t2 := New(), New()
	
	var val int32 = 234234
	
	// Encode in the value.
	t1.Put(&val)
	
	// Copy to the other.
	t2.Write(t1.Read())
	
	// Verify accessors work.
	var ta1, ta2 *int32
	if ta1 = Assert(t1); ta1 == nil {
		log.E.Ln("did not get expected time value")
		t.FailNow()
	}
	if ta2 = Assert(t2); ta2 == nil {
		log.E.Ln("did not get expected time value")
		t.FailNow()
	}
	
	// Verify the value survived the encode/decode.
	if *ta1 != *ta2 {
		t.FailNow()
	}
	
	// Test NewFrom correctly decodes and returns the trimmings.
	b1 := t1.Read()
	nb1, rem := NewFrom(append(b1, make([]byte, 5)...))
	if len(rem) != 5 || rem == nil {
		t.FailNow()
	}
	v := Assert(nb1)
	if *v != *ta1 {
		t.FailNow()
	}
}
