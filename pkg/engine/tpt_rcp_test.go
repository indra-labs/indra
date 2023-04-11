package engine

import (
	"context"
	"testing"
	"time"
	
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
)

const localhostZeroIPv4 = "/ip4/127.0.0.1/tcp/0"

func TestNewRCPListener(t *testing.T) {
	log2.SetLogLevel(log2.Trace)
	var e error
	var l1, l2 *RCPListener
	_ = l2
	var k1, k2 *Keys
	ctx, cancel := context.WithCancel(context.Background())
	_ = cancel
	if k1, k2, e = Generate2Keys(); fails(e) {
		t.FailNow()
	}
	l1, e = NewRCPListener("", localhostZeroIPv4, k1.Prv, ctx)
	if fails(e) {
		t.FailNow()
	}
	l2, e = NewRCPListener(getHostAddress(l1.Host), localhostZeroIPv4,
		k2.Prv, ctx)
	if fails(e) {
		t.FailNow()
	}
	time.Sleep(time.Second)
	
	time.Sleep(time.Second)
	cancel()
}
