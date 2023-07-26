package transport

import (
	"context"
	"github.com/indra-labs/indra"
	"os"
	"testing"
	"time"

	"github.com/indra-labs/indra/pkg/crypto"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/tests"
)

func TestNewListener(t *testing.T) {
	if indra.CI == "false" {
		log2.SetLogLevel(log2.Trace)
	}
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
	store, closer := BadgerStore(dataPath)
	if store == nil {
		t.Fatal("could not open database")
	}
	l1, e = NewListener([]string{""}, []string{LocalhostZeroIPv4TCP}, k1, store,
		closer, ctx, DefaultMTU, cancel)
	if fails(e) {
		t.FailNow()
	}
	dataPath, err = os.MkdirTemp(os.TempDir(), "badger")
	if err != nil {
		t.FailNow()
	}
	store, closer = BadgerStore(dataPath)
	if store == nil {
		t.Fatal("could not open database")
	}
	l2, e = NewListener([]string{GetHostFirstMultiaddr(l1.Host)},
		[]string{LocalhostZeroIPv4TCP}, k2, store, closer, ctx, DefaultMTU,
		cancel)
	if fails(e) {
		t.FailNow()
	}
	var msg1, msg2 []byte
	_ = msg2
	msg1, _, e = tests.GenMessage(32, "REQUEST")
	msg2, _, e = tests.GenMessage(32, "RESPONSE")
	hn1 := GetHostFirstMultiaddr(l2.Host)
	hn2 := GetHostFirstMultiaddr(l1.Host)
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
