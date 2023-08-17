package time

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

func TestStamp(t *testing.T) {
	
	if indra.CI == "false" {
		log2.SetLogLevel(log2.Trace)
	}
	
	t1 := New().Put(time.Now())
	t2 := New()
	t2.Decode(t1.Encode())
	if t1.Get() != t2.Get() {
		t.Fail()
	}
}
