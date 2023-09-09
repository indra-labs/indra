package u32

import (
	log2 "git.indra-labs.org/dev/ind/pkg/proc/log"
	"git.indra-labs.org/dev/ind/pkg/util/ci"
	"math"
	"math/rand"
	"testing"
)

var (
	log   = log2.GetLogger()
	fails = log.E.Chk
)

func TestU(t *testing.T) {
	ci.TraceIfNot()
	
	t1, t2 := New(), New()
	
	var val = uint32(rand.Intn(math.MaxInt32))
	
	// Encode in the value.
	t1.Put(&val)
	
	// Copy to the other.
	t2.Write(t1.Read())
	
	// Verify accessors work.
	var ta1, ta2 *uint32
	if ta1 = Assert(t1); ta1 == nil {
		log.E.Ln("did not get expected value")
		t.FailNow()
	}
	if ta2 = Assert(t2); ta2 == nil {
		log.E.Ln("did not get expected value")
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
