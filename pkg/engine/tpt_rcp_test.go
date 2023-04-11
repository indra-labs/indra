package engine

import (
	"context"
	"testing"
	"time"
	
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/util/tests"
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
	var msg1, msg2 []byte
	_ = msg2
	msg1, _, e = tests.GenMessage(32, "REQUEST")
	msg2, _, e = tests.GenMessage(32, "RESPONSE")
	hn1 := getHostAddress(l2.Host)
	hn2 := getHostAddress(l1.Host)
	d1 := l1.DialRCP(hn1)
	d2 := l2.DialRCP(hn2)
	l1.Lock()
	l2.Lock()
	c1, c2 := l1.Connections[hn1].Recv, l2.Connections[hn2].Recv
	l1.Unlock()
	l2.Unlock()
	go func() {
		for {
			select {
			case b := <-c1:
				log.D.S("received "+hn1, b.ToBytes())
				// d1.Send <- msg2
			case b := <-c2:
				log.D.S("received "+hn2, b.ToBytes())
				d2.Send <- msg2
			case <-ctx.Done():
				return
			}
		}
	}()
	time.Sleep(time.Second)
	l1.Lock()
	l2.Lock()
	log.D.Ln("connections", l1.Connections, l2.Connections)
	l1.Unlock()
	l2.Unlock()
	d1.Send <- msg1
	d2.Send <- msg1
	time.Sleep(time.Second)
	d1.Send <- msg1
	d2.Send <- msg1
	time.Sleep(time.Second)
	cancel()
}
