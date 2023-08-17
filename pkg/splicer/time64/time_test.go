package time64

import (
	indra "git.indra-labs.org/dev/ind"
	log2 "git.indra-labs.org/dev/ind/pkg/proc/log"
	"testing"
	"time"
)

var (
	log   = log2.GetLogger()
	fails = log.E.Chk
)

func TestNew(t *testing.T) {
	if indra.CI == "false" {
		log2.SetLogLevel(log2.Trace)
	}
	
	t1, t2 := New(), New()
	nao := time.Now()
	
	// Encode in the time value.
	t1.Put(&nao)
	
	// Copy to other Stamp.
	t2.Write(t1.Read())
	
	// Verify accessors work.
	var ta1, ta2 *time.Time
	if ta1 = Assert(t1); ta1 == nil {
		log.E.Ln("did not get expected time value")
		t.FailNow()
	}
	if ta2 = Assert(t2); ta2 == nil {
		log.E.Ln("did not get expected time value")
		t.FailNow()
	}
	
	// Verify the value survived the encode/decode.
	if !(*ta1).Equal(*ta2) {
		t.FailNow()
	}
	
	// Test NewFrom correctly decodes and returns the trimmings.
	b1 := t1.Read()
	nb1, rem := NewFrom(append(b1, make([]byte, 5)...))
	if len(rem) != 5 || rem == nil {
		t.FailNow()
	}
	val := Assert(nb1)
	if !(*val).Equal(*ta1) {
		t.FailNow()
	}
}
