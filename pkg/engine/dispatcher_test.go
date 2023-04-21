package engine

import (
	"context"
	"testing"
	"time"
	
	"git-indra.lan/indra-labs/indra/pkg/engine/transport"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/util/tests"
)

func TestDispatcher(t *testing.T) {
	log2.SetLogLevel(log2.Trace)
	var e error
	var l1, l2 *transport.Listener
	_ = l2
	var k1, k2 *crypto.Keys
	ctx, cancel := context.WithCancel(context.Background())
	_ = cancel
	if k1, k2, e = crypto.Generate2Keys(); fails(e) {
		t.FailNow()
	}
	l1, e = transport.NewListener("", transport.LocalhostZeroIPv4, k1.Prv, ctx,
		transport.DefaultMTU)
	if fails(e) {
		t.FailNow()
	}
	l2, e = transport.NewListener(transport.GetHostAddress(l1.Host),
		transport.LocalhostZeroIPv4,
		k2.Prv, ctx, transport.DefaultMTU)
	if fails(e) {
		t.FailNow()
	}
	var msg1, msg2 []byte
	_ = msg2
	msg1, _, e = tests.GenMessage(48, "REQUEST")
	msg2, _, e = tests.GenMessage(48, "RESPONSE")
	_, _ = msg1, msg2
	hn1 := transport.GetHostAddress(l2.Host)
	hn2 := transport.GetHostAddress(l1.Host)
	var ks *crypto.KeySet
	_, ks, e = crypto.NewSigner()
	d1 := NewDispatcher(l1.Dial(hn1), k1.Prv, ctx, ks)
	d2 := NewDispatcher(l2.Dial(hn2), k2.Prv, ctx, ks)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case b := <-d1.Duplex.Receive():
				log.W.S("duplex receive "+d1.Conn.ID(), b.ToBytes())
			case b := <-d2.Duplex.Receive():
				log.W.S("duplex receive "+d2.Conn.ID(), b.ToBytes())
			}
		}
	}()
	var msgp1, msgp2 slice.Bytes
	_ = msgp2
	_ = msgp1
	
	d1.Duplex.Send(msg1)
	// d1.SendToConn(msgp1)
	// d2.SendToConn(msg2)
	time.Sleep(time.Second * 3)
	log.D.Ln("ping", time.Duration(d1.Ping.Value()),
		time.Duration(d2.Ping.Value()))
	cancel()
}
