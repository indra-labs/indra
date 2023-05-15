package transport

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/util/appdata"
	"git-indra.lan/indra-labs/indra/pkg/util/tests"
)

func TestNewListener(t *testing.T) {
	log2.SetLogLevel(log2.Debug)
	var e error
	var l1, l2 *Listener
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
	l1, e = NewListener("", LocalhostZeroIPv4TCP, dataPath, k1, ctx, DefaultMTU)
	if fails(e) {
		t.FailNow()
	}
	l2, e = NewListener(GetHostAddress(l1.Host), LocalhostZeroIPv4TCP,
		filepath.Join(appdata.MakeDirIfNeeded("indra", false, 0770), "badgerdb"), k2, ctx, DefaultMTU)
	if fails(e) {
		t.FailNow()
	}
	var msg1, msg2 []byte
	_ = msg2
	msg1, _, e = tests.GenMessage(32, "REQUEST")
	msg2, _, e = tests.GenMessage(32, "RESPONSE")
	hn1 := GetHostAddress(l2.Host)
	hn2 := GetHostAddress(l1.Host)
	d1 := l1.Dial(hn1)
	d2 := l2.Dial(hn2)
	c1, c2 := l1.GetConnRecv(hn1), l2.GetConnRecv(hn2)
	go func() {
		for {
			select {
			case b := <-c1.Receive():
				log.D.S("received "+hn1, b.ToBytes())
			case b := <-c2.Receive():
				log.D.S("received "+hn2, b.ToBytes())
				d2.Transport.Sender.Send(msg2)
			case <-ctx.Done():
				return
			}
		}
	}()
	time.Sleep(time.Second)
	l1.Lock()
	l2.Lock()
	log.D.Ln("connections", l1.connections, l2.connections)
	l1.Unlock()
	l2.Unlock()
	d1.Transport.Sender.Send(msg1)
	d2.Transport.Sender.Send(msg1)
	time.Sleep(time.Second)
	d1.Transport.Sender.Send(msg1)
	d2.Transport.Sender.Send(msg1)
	time.Sleep(time.Second)
	cancel()
}

func TestDHT(t *testing.T) {
	log2.SetLogLevel(log2.Debug)
	var e error
	const nTotal = 26
	ctx, cancel := context.WithCancel(context.Background())
	var listeners []*Listener
	var keys []*crypto.Keys
	var seed string
	for i := 0; i < nTotal; i++ {
		var k *crypto.Keys
		if k, e = crypto.GenerateKeys(); fails(e) {
			t.FailNow()
		}
		keys = append(keys, k)
		var l *Listener
		dataPath, err := os.MkdirTemp(os.TempDir(), "badger")
		if err != nil {
		}
		// filepath.Join(appdata.MakeDirIfNeeded("indra", false, 0770))
		if l, e = NewListener(seed, LocalhostZeroIPv4TCP, dataPath, k, ctx,
			DefaultMTU); fails(e) {
			t.FailNow()
		}
		sa := GetHostAddress(l.Host)
		if i == 0 {
			seed = sa
		}
		listeners = append(listeners, l)
	}
	for i, l := range listeners {
		_, _ = i, l
		// rc := record.MakePutRecord(string(l.Keys.Bytes[:]),
		// 	)
		// if e = l.DHT.PutValue(ctx, ProtocolPrefix+"/ns/a/"+l.Keys.Bytes.
		// 	String(), []byte{}); fails(e) {
		//
		// 	t.FailNow()
		// }
	}
	// listeners[0].DHT.PutValue(ctx, )
	time.Sleep(time.Second * 2)
	cancel()
}
