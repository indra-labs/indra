package engine

import (
	"testing"
	
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
)

const localhostZeroIPv4 = "/ip4/127.0.0.1/tcp/0"

func TestNewRCPListener(t *testing.T) {
	log2.SetLogLevel(log2.Trace)
	var e error
	var l *RCPListener
	var k *Keys
	if k, e = GenerateKeys(); fails(e) {
		t.FailNow()
	}
	l, e = NewRCPListener(localhostZeroIPv4, k.Prv, 1440, 32)
	if fails(e) {
		t.FailNow()
	}
	hi := l.Host.ID().String()
	log.D.S("lis", hi, l.Host.Addrs(), k.Pub.ToBytes())
	
}
