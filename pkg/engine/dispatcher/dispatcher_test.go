package dispatcher

import (
	"context"
	"github.com/indra-labs/indra/pkg/onions"
	"github.com/indra-labs/indra/pkg/onions/confirmation"
	"github.com/indra-labs/indra/pkg/onions/ont"
	"github.com/indra-labs/indra/pkg/onions/response"
	"os"
	"testing"
	"time"

	"github.com/indra-labs/indra/pkg/util/slice"
	"github.com/indra-labs/indra/pkg/util/splice"

	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/engine/transport"

	"github.com/indra-labs/indra/pkg/crypto"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/tests"
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
	on1 := ont.Assemble(onions.Skins{
		confirmation.NewConfirmation(id1, load1)})
	on2 := ont.Assemble(onions.Skins{
		response.NewResponse(id2, 0, msg1, 0)})
	s1 := ont.Encode(on1)
	s2 := ont.Encode(on2)
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
	msg1, _, e = tests.GenMessage(4096, "REQUEST")
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
	on1 := ont.Assemble(onions.Skins{
		response.NewResponse(id1, 0, msg1, 0)})
	on2 := ont.Assemble(onions.Skins{
		response.NewResponse(id2, 0, msg2, 0)})
	s1 := ont.Encode(on1)
	s2 := ont.Encode(on2)
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
					continue
				}
			case b := <-d2.Duplex.Receive():
				bb, xb1 := b.ToBytes(), x1.ToBytes()
				if string(bb) != string(xb1) {
					t.Error("did not receive expected message")
					return
				} else {
					succ++
					continue
				}
			}
		}
	}()
	msgp1 = sp1.GetAll()
	msgp2 = sp2.GetAll()
	for i := 0; i < countTo; i++ {
		d1.SendToConn(msgp1)
		d2.SendToConn(msgp2)
	}
	time.Sleep(time.Second)
	cancel()
	if succ != countTo*3 {
		t.Fatal("did not receive all messages correctly", succ, countTo*3)
	}
}
