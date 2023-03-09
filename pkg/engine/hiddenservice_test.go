package engine

import (
	"sync"
	"testing"
	"time"
	
	"github.com/cybriq/qu"
	"go.uber.org/atomic"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/signer"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

func TestOnionSkins_HiddenService(t *testing.T) {
	var e error
	n3 := Gen3Nonces()
	id := nonce.NewID()
	pr, ks, _ := signer.New()
	in := NewIntro(pr, slice.GenerateRandomAddrPortIPv6())
	var prvs [3]*prv.Key
	for i := range prvs {
		prvs[i] = ks.Next()
	}
	var pubs [3]*pub.Key
	for i := range pubs {
		pubs[i] = pub.Derive(prvs[i])
	}
	on1 := Skins{}.
		HiddenService(id, in, prvs, pubs, n3)
	on1 = append(on1, &Tmpl{})
	on := on1.Assemble()
	s := Encode(on)
	s.SetCursor(0)
	var onc Onion
	if onc = Recognise(s); onc == nil {
		t.Error("did not unwrap")
		t.FailNow()
	}
	if e = onc.Decode(s); check(e) {
		t.Error("did not decode")
		t.FailNow()
	}
	var hs *HiddenService
	var ok bool
	if hs, ok = onc.(*HiddenService); !ok {
		t.Error("did not unwrap expected type")
		t.FailNow()
	}
	if hs.ID != id {
		t.Error("ID did not decode correctly")
		t.FailNow()
	}
	for i := range hs.Ciphers {
		if hs.Ciphers[i] != on.(*HiddenService).Ciphers[i] {
			t.Errorf("cipher %d did not unwrap correctly", i)
			t.FailNow()
		}
	}
	for i := range hs.Nonces {
		if hs.Nonces[i] != n3[i] {
			t.Errorf("nonce %d did not unwrap correctly", i)
			t.FailNow()
		}
	}
	if !hs.Intro.Key.Equals(in.Key) {
		t.Errorf("key did not decode correctly")
		t.FailNow()
	}
	if hs.AddrPort.String() != in.AddrPort.String() {
		t.Errorf("addrport did not decode correctly")
		t.FailNow()
	}
	if string(hs.Intro.Sig[:]) != string(in.Sig[:]) {
		t.Errorf("signature did not decode correctly")
		t.FailNow()
	}
	if !hs.Intro.Validate() {
		t.Errorf("received intro did not validate")
		t.FailNow()
	}
}

func TestEngine_MakeHiddenService(t *testing.T) {
	log2.SetLogLevel(log2.Debug)
	log2.App = "test"
	var clients []*Engine
	var e error
	const nCircuits = 10
	if clients, e = CreateNMockCircuits(nCircuits, nCircuits); check(e) {
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
		case <-quit:
			for _, v := range clients {
				v.Shutdown()
			}
			quit.Q()
			return
		}
		for i := 0; i < int(counter.Load()); i++ {
			wg.Done()
		}
		t.Error("MakeHiddenService test failed")
	}()
	for i := 0; i < nCircuits*nCircuits/2; i++ {
		wg.Add(1)
		counter.Inc()
		e = clients[0].BuyNewSessions(1000000, func() {
			wg.Done()
			counter.Dec()
		})
		if check(e) {
			wg.Done()
			counter.Dec()
		}
		wg.Wait()
	}
	
	quit.Q()
}
