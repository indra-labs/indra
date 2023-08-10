package dispatcher

import (
	"context"
	"crypto/rand"
	"git.indra-labs.org/dev/ind/pkg/codec"
	"git.indra-labs.org/dev/ind/pkg/codec/onion/cores/confirmation"
	"git.indra-labs.org/dev/ind/pkg/codec/onion/cores/response"
	"git.indra-labs.org/dev/ind/pkg/crypto/sha256"
	"os"
	"testing"
	"time"

	"git.indra-labs.org/dev/ind"
	"git.indra-labs.org/dev/ind/pkg/codec/ont"
	"git.indra-labs.org/dev/ind/pkg/engine"

	"git.indra-labs.org/dev/ind/pkg/util/slice"
	"git.indra-labs.org/dev/ind/pkg/util/splice"

	"git.indra-labs.org/dev/ind/pkg/crypto/nonce"
	"git.indra-labs.org/dev/ind/pkg/engine/transport"

	"git.indra-labs.org/dev/ind/pkg/crypto"
	log2 "git.indra-labs.org/dev/ind/pkg/proc/log"
	"git.indra-labs.org/dev/ind/pkg/util/tests"
)

func TestDispatcher(t *testing.T) {
	t.Log(indra.CI)
	if indra.CI == "false" {
		log2.SetLogLevel(log2.Trace)
		log.D.Ln("debug")
	}
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
	secret := sha256.New()
	rand.Read(secret[:])
	store, closer := transport.BadgerStore(dataPath, secret[:])
	if store == nil {
		t.Fatal("could not open database")
	}
	l1, e = transport.NewListener([]string{""},
		[]string{transport.LocalhostZeroIPv4TCP},
		k1, store, closer, ctx, transport.DefaultMTU, cancel)
	if fails(e) {
		t.FailNow()
	}
	dataPath, err = os.MkdirTemp(os.TempDir(), "badger")
	if err != nil {
		t.FailNow()
	}
	secret = sha256.New()
	rand.Read(secret[:])
	store, closer = transport.BadgerStore(dataPath, secret[:])
	if store == nil {
		t.Fatal("could not open database")
	}
	l1, e = transport.NewListener([]string{""},
		[]string{transport.LocalhostZeroIPv4TCP},
		k1, store, closer, ctx, transport.DefaultMTU, cancel)
	if fails(e) {
		t.FailNow()
	}
	l2, e = transport.NewListener([]string{transport.GetHostFirstMultiaddr(l1.Host)},
		[]string{transport.LocalhostZeroIPv4TCP}, k2, store, closer, ctx,
		transport.DefaultMTU, cancel)
	if fails(e) {
		t.FailNow()
	}
	var msg1, msg2 []byte
	_ = msg2
	msg1, _, e = tests.GenMessage(8192, "REQUEST")
	msg2, _, e = tests.GenMessage(4096, "RESPONSE")
	_, _ = msg1, msg2
	hn1 := transport.GetHostFirstMultiaddr(l2.Host)
	// hn2 := transport.GetHostAddress(l1.Host)
	var ks *crypto.KeySet
	_, ks, e = crypto.NewSigner()
	d1 := NewDispatcher(l1.Dial(hn1), ctx, ks)
	d2 := NewDispatcher(<-l2.Accept(), ctx, ks)
	var msgp1, msgp2 slice.Bytes
	id1, id2 := nonce.NewID(), nonce.NewID()
	// var load2 byte = 32
	on1 := ont.Assemble(engine.Skins{
		confirmation.New(id1)})
	on2 := ont.Assemble(engine.Skins{
		response.New(id2, 0, msg1, 0)})
	s1 := codec.Encode(on1)
	s2 := codec.Encode(on2)
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
					log.D.Ln("success", count)
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
					log.D.Ln("success", count)
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
	log.D.Ln("sent 1")
	time.Sleep(time.Second)
	d2.SendToConn(msgp2)
	// wg.Add(n)
	log.D.Ln("sent 2")
	time.Sleep(time.Second)
	d1.SendToConn(msgp1)
	// wg.Add(n)
	log.D.Ln("sent 3")
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
