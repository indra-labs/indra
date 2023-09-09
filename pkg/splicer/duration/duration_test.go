package duration

import (
	log2 "git.indra-labs.org/dev/ind/pkg/proc/log"
	"git.indra-labs.org/dev/ind/pkg/util/ci"
	"testing"
	"time"
)

var (
	log   = log2.GetLogger()
	fails = log.E.Chk
)

func TestNew(t *testing.T) {
	ci.TraceIfNot()
	
	t1, t2 := New(), New()
	nao, e := time.ParseDuration("21h33m11s")
	
	if e != nil {
		t.Fatal(e)
	}
	
	// Encode in the time value.
	t1.Put(&nao)
	
	// Copy to other T.
	t2.Write(t1.Read())
	
	// Verify accessors work.
	var ta1, ta2 *time.Duration
	if ta1 = Assert(t1); ta1 == nil {
		log.E.Ln("did not get expected duration value")
		t.FailNow()
	}
	if ta2 = Assert(t2); ta2 == nil {
		log.E.Ln("did not get expected duration value")
		t.FailNow()
	}
	
	// Verify the value survived the encode/decode.
	if *ta1 != *ta2 {
		t.FailNow()
	}
	
	// Test NewFrom correctly decodes and returns the trimmings.
	b1 := t1.Read()
	nb1, rem := NewFrom(append(b1, make([]byte, 5)...))
	if rem == nil || len(rem) != 5 {
		t.FailNow()
	}
	val := Assert(nb1)
	if *val != *ta1 {
		t.FailNow()
	}
}
