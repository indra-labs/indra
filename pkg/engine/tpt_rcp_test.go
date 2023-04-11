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
	d1 := l1.DialRCP(getHostAddress(l2.Host))
	d2 := l2.DialRCP(getHostAddress(l1.Host))
	go func() {
		for {
			log.D.Ln("recv loop")
			select {
			case b := <-d1.Recv:
				log.D.S("d1", b)
			case b := <-d2.Recv:
				log.D.S("d2", b)
			case <-ctx.Done():
				break
			}
		}
	}()
	time.Sleep(time.Second)
	go func() {
		for {
			select {
			case b := <-l1.MsgChan:
				log.D.S(getHostAddress(l1.Host)+" received", b.sender,
					b.ToBytes())
				if b.String() != string(msg2) {
					d1.Send <- msg2
				}
			case b := <-l2.MsgChan:
				log.D.S(getHostAddress(l2.Host)+" received", b.sender,
					b.ToBytes())
				if b.String() != string(msg2) {
					d2.Send <- msg2
				}
			case <-ctx.Done():
				break
			}
		}
	}()
	d1.Send <- msg1
	d2.Send <- msg1
	time.Sleep(time.Second)
	d1.Send <- msg1
	d2.Send <- msg1
	time.Sleep(time.Second)
	cancel()
}
