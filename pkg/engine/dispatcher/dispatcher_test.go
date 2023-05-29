package dispatcher

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"git-indra.lan/indra-labs/indra/pkg/engine/onions"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
	"git-indra.lan/indra-labs/indra/pkg/util/splice"

	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/engine/transport"

	"git-indra.lan/indra-labs/indra/pkg/crypto"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/util/tests"
)

func TestDispatcher(t *testing.T) {
	log2.SetLogLevel(log2.Debug)
	var e error
	var l1, l2 *transport.Listener
	_ = l2
	var k1, k2 *crypto.Keys
	ctx, cancel := context.WithCancel(context.Background())
	_ = cancel
	if k1, k2, e = crypto.Generate2Keys(); fails(e) {
		t.FailNow()
	}
	dataPath, err := os.MkdirTemp(os.TempDir(), "badger")
	if err != nil {
		t.FailNow()
	}
	l1, e = transport.NewListener("", transport.LocalhostZeroIPv4QUIC,
		dataPath, k1, ctx, transport.DefaultMTU)
	if fails(e) {
		t.FailNow()
	}
	dataPath, err = os.MkdirTemp(os.TempDir(), "badger")
	if err != nil {
		t.FailNow()
	}
	l2, e = transport.NewListener(transport.GetHostAddress(l1.Host),
		transport.LocalhostZeroIPv4QUIC, dataPath, k2, ctx, transport.DefaultMTU)
	if fails(e) {
		t.FailNow()
	}
	var msg1, msg2 []byte
	_ = msg2
	msg1, _, e = tests.GenMessage(8192, "REQUEST")
	msg2, _, e = tests.GenMessage(4096, "RESPONSE")
	_, _ = msg1, msg2
	hn1 := transport.GetHostAddress(l2.Host)
	// hn2 := transport.GetHostAddress(l1.Host)
	var ks *crypto.KeySet
	_, ks, e = crypto.NewSigner()
	d1 := NewDispatcher(l1.Dial(hn1), ctx, ks)
	d2 := NewDispatcher(<-l2.Accept(), ctx, ks)
	var msgp1, msgp2 slice.Bytes
	id1, id2 := nonce.NewID(), nonce.NewID()
	var load1 byte = 128
	// var load2 byte = 32
	on1 := onions.Skins{}.
		Confirmation(id1, load1).
		Assemble()
	on2 := onions.Skins{}.
		Response(id2, msg1, 0).
		Assemble()
	s1 := onions.Encode(on1)
	s2 := onions.Encode(on2)
	x1 := s1.GetAll()
	x2 := s2.GetAll()
	xx1 := &Onion{x1}
	xx2 := &Onion{x2}
	sp1 := splice.New(xx1.Len())
	sp2 := splice.New(xx2.Len())
	if e = xx1.Encode(sp1); fails(e) {
		t.FailNow()
	}
	if e = xx2.Encode(sp2); fails(e) {
		t.FailNow()
	}
	// var wg sync.WaitGroup
	go func() {
		var count int
		for {
			select {
			case <-ctx.Done():
				return
			case b := <-d1.Duplex.Receive():
				bb, xb2 := b.ToBytes(), x2.ToBytes()
				if string(bb) != string(xb2) {
					t.Error("did not receive expected message")
					return
				} else {
					log.I.Ln("success", count)
					count++
					// wg.Done()
					continue
				}
			case b := <-d2.Duplex.Receive():
				bb, xb1 := b.ToBytes(), x1.ToBytes()
				if string(bb) != string(xb1) {
					t.Error("did not receive expected message")
					return
				} else {
					log.I.Ln("success", count)
					count++
					// wg.Done()
					continue
				}
			}
		}
	}()
	msgp1 = sp1.GetAll()
	msgp2 = sp2.GetAll()
	time.Sleep(time.Second)
	// var n int
	d1.SendToConn(msgp1)
	// wg.Add(n)
	log.I.Ln("sent 1")
	time.Sleep(time.Second)
	d2.SendToConn(msgp2)
	// wg.Add(n)
	log.I.Ln("sent 2")
	time.Sleep(time.Second)
	d1.SendToConn(msgp1)
	// wg.Add(n)
	log.I.Ln("sent 3")
	time.Sleep(time.Second)
	d2.SendToConn(msgp2)
	// wg.Add(n)
	// wg.Wait()
	time.Sleep(time.Second)
	d1.Mutex.Lock()
	d2.Mutex.Lock()
	log.D.Ln("ping", time.Duration(d1.Ping.Value()),
		time.Duration(d2.Ping.Value()))
	d1.Mutex.Unlock()
	d2.Mutex.Unlock()
	cancel()
}

func TestDispatcher_Rekey(t *testing.T) {
	log2.SetLogLevel(log2.Debug)
	var e error
	var l1, l2 *transport.Listener
	_ = l2
	var k1, k2 *crypto.Keys
	ctx, cancel := context.WithCancel(context.Background())
	_ = cancel
	if k1, k2, e = crypto.Generate2Keys(); fails(e) {
		t.FailNow()
	}
	dataPath, err := os.MkdirTemp(os.TempDir(), "badger")
	if err != nil {
		t.FailNow()
	}
	l1, e = transport.NewListener("", transport.LocalhostZeroIPv4TCP,
		dataPath, k1, ctx, transport.DefaultMTU)
	if fails(e) {
		t.FailNow()
	}
	dataPath, err = os.MkdirTemp(os.TempDir(), "badger")
	if err != nil {
		t.FailNow()
	}
	l2, e = transport.NewListener(transport.GetHostAddress(l1.Host),
		transport.LocalhostZeroIPv4TCP, dataPath, k2, ctx, transport.DefaultMTU)
	if fails(e) {
		t.FailNow()
	}
	var msg1, msg2 []byte
	_ = msg2
	msg1, _, e = tests.GenMessage(1024, "REQUEST")
	msg2, _, e = tests.GenMessage(1024, "RESPONSE")
	_, _ = msg1, msg2
	hn1 := transport.GetHostAddress(l2.Host)
	// hn2 := transport.GetHostAddress(l1.Host)
	var ks *crypto.KeySet
	_, ks, e = crypto.NewSigner()
	d1 := NewDispatcher(l1.Dial(hn1), ctx, ks)
	d2 := NewDispatcher(<-l2.Accept(), ctx, ks)
	_, _ = d1, d2
	var msgp1, msgp2 slice.Bytes
	id1, id2 := nonce.NewID(), nonce.NewID()
	on1 := onions.Skins{}.
		Response(id1, msg1, 0).
		Assemble()
	on2 := onions.Skins{}.
		Response(id2, msg2, 0).
		Assemble()
	s1 := onions.Encode(on1)
	s2 := onions.Encode(on2)
	x1 := s1.GetAll()
	x2 := s2.GetAll()
	xx1 := &Onion{x1}
	xx2 := &Onion{x2}
	sp1 := splice.New(xx1.Len())
	sp2 := splice.New(xx2.Len())
	if e = xx1.Encode(sp1); fails(e) {
		t.FailNow()
	}
	if e = xx2.Encode(sp2); fails(e) {
		t.FailNow()
	}
	countTo, succ := 1000, 0
	var wg sync.WaitGroup
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case b := <-d1.Duplex.Receive():
				bb, xb2 := b.ToBytes(), x2.ToBytes()
				if string(bb) != string(xb2) {
					t.Error("did not receive expected message")
					return
				} else {
					succ++
					wg.Done()
					continue
				}
			case b := <-d2.Duplex.Receive():
				bb, xb1 := b.ToBytes(), x1.ToBytes()
				if string(bb) != string(xb1) {
					t.Error("did not receive expected message")
					return
				} else {
					succ++
					wg.Done()
					continue
				}
			}
		}
	}()
	msgp1 = sp1.GetAll()
	msgp2 = sp2.GetAll()
	for i := 0; i < countTo; i++ {
		wg.Add(d1.SendToConn(msgp1))
		wg.Add(d2.SendToConn(msgp2))
		wg.Wait()
	}
	cancel()
	if succ != countTo*2 {
		t.Fatal("did not receive all messages correctly")
	}
}
