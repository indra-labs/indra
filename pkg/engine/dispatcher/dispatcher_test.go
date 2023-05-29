package dispatcher

import (
	"context"
	"git-indra.lan/indra-labs/indra/pkg/engine"
	"git-indra.lan/indra-labs/indra/pkg/engine/services"
	"git-indra.lan/indra-labs/indra/pkg/engine/sessions"
	"git-indra.lan/indra-labs/indra/pkg/util/cryptorand"
	"git-indra.lan/indra-labs/indra/pkg/util/qu"
	"go.uber.org/atomic"
	"os"
	"reflect"
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

func TestEngine_SendHiddenService(t *testing.T) {
	log2.SetLogLevel(log2.Debug)
	log2.App.Store("")
	var clients []*engine.Engine
	var e error
	const nCircuits = 10
	ctx, cancel := context.WithCancel(context.Background())
	if clients, e = engine.CreateNMockCircuits(nCircuits, nCircuits,
		ctx); fails(e) {
		t.Error(e)
		t.FailNow()
	}
	// Start up the clients.
	for _, v := range clients {
		go v.Start()
	}
	var wg sync.WaitGroup
	var counter atomic.Int32
	quit := qu.T()
	go func() {
		select {
		case <-time.After(time.Second * 6):
			quit.Q()
			t.Error("MakeHiddenService test failed")
		case <-quit:
			for _, v := range clients {
				v.Shutdown()
			}
			for i := 0; i < int(counter.Load()); i++ {
				wg.Done()
			}
			return
		}
	}()
	for i := 0; i < nCircuits*nCircuits/2; i++ {
		wg.Add(1)
		counter.Inc()
		e = clients[0].BuyNewSessions(1000000, func() {
			wg.Done()
			counter.Dec()
		})
		if fails(e) {
			wg.Done()
			counter.Dec()
		}
		wg.Wait()
	}
	log2.SetLogLevel(log2.Debug)
	var idPrv *crypto.Prv
	if idPrv, e = crypto.GeneratePrvKey(); fails(e) {
		return
	}
	id := nonce.NewID()
	introducerHops := clients[0].Manager.GetSessionsAtHop(2)
	returnHops := clients[0].Manager.GetSessionsAtHop(5)
	var introducer *sessions.Data
	if len(introducerHops) > 1 {
		cryptorand.Shuffle(len(introducerHops), func(i, j int) {
			introducerHops[i], introducerHops[j] = introducerHops[j],
				introducerHops[i]
		})
	}
	var returner *sessions.Data
	if len(introducerHops) > 1 {
		cryptorand.Shuffle(len(returnHops), func(i, j int) {
			returnHops[i], returnHops[j] = returnHops[j],
				returnHops[i]
		})
	}
	// There must be at least one, and if there was more than one the first
	// index of introducerHops will be a randomly selected one.
	introducer = introducerHops[0]
	returner = returnHops[0]
	wg.Add(1)
	counter.Inc()
	svc := &services.Service{
		Port:      2345,
		RelayRate: 43523,
		Transport: transport.NewByteChan(64),
	}
	clients[0].SendHiddenService(id, idPrv, 0, 0,
		time.Now().Add(time.Hour), returner, introducer, svc,
		func(id nonce.ID, ifc interface{}, b slice.Bytes) (e error) {
			log.W.S("received intro", reflect.TypeOf(ifc), b.ToBytes())
			// This happens when the gossip gets back to us.
			wg.Done()
			counter.Dec()
			return
		})
	wg.Wait()
	quit.Q()
	cancel()
}
