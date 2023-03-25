package engine

import (
	"sync"
	"testing"
	"time"
	
	"github.com/cybriq/qu"
	"go.uber.org/atomic"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	log2 "git-indra.lan/indra-labs/indra/pkg/proc/log"
	"git-indra.lan/indra-labs/indra/pkg/util/cryptorand"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

func TestOnionSkins_IntroQuery(t *testing.T) {
	log2.SetLogLevel(log2.Trace)
	log2.App = ""
	var e error
	prvs, pubs := GetCipherSet(t)
	ciphers := GenCiphers(prvs, pubs)
	prv1, _ := GetTwoPrvKeys(t)
	pub1 := pub.Derive(prv1)
	n3 := Gen3Nonces()
	ep := &ExitPoint{
		Routing: &Routing{
			Sessions: [3]*SessionData{},
			Keys:     prvs,
			Nonces:   n3,
		},
		ReturnPubs: pubs,
	}
	id := nonce.NewID()
	on := Skins{}.
		IntroQuery(id, pub.Derive(prv1), ep).
		End().Assemble()
	s := Encode(on)
	s.SetCursor(0)
	var onc Onion
	if onc = Recognise(s, slice.GenerateRandomAddrPortIPv6()); onc == nil {
		t.Error("did not unwrap")
		t.FailNow()
	}
	if e = onc.Decode(s); check(e) {
		t.Error("did not decode")
		t.FailNow()
	}
	log.D.Ln(s)
	var ex *IntroQuery
	var ok bool
	if ex, ok = onc.(*IntroQuery); !ok {
		t.Error("did not unwrap expected type")
		t.FailNow()
	}
	for i := range ex.Ciphers {
		if ex.Ciphers[i] != ciphers[i] {
			t.Errorf("cipher %d did not unwrap correctly", i)
			t.FailNow()
		}
	}
	for i := range ex.Nonces {
		if ex.Nonces[i] != n3[i] {
			t.Errorf("nonce %d did not unwrap correctly", i)
			t.FailNow()
		}
	}
	if !ex.Key.Equals(pub1) {
		t.Error("HiddenService did not decode correctly")
		t.FailNow()
	}
	if ex.ID != id {
		t.Error("ID did not decode correctly")
		t.FailNow()
	}
}

func TestEngine_SendIntroQuery(t *testing.T) {
	log2.SetLogLevel(log2.Info)
	log2.App = ""
	var clients []*Engine
	var e error
	const nCircuits = 10
	if clients, e = CreateNMockCircuits(nCircuits, nCircuits); check(e) {
		t.Error(e)
		t.FailNow()
	}
	client := clients[0]
	// Start up the clients.
	for _, v := range clients {
		go v.Start()
	}
	var wg sync.WaitGroup
	var counter atomic.Int32
	quit := qu.T()
	go func() {
		for {
			select {
			case <-time.After(time.Second * 4):
				quit.Q()
				t.Error("MakeHiddenService test failed")
			case <-quit:
				for i := 0; i < int(counter.Load()); i++ {
					wg.Done()
				}
				for _, v := range clients {
					v.Shutdown()
				}
				return
			}
		}
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
	var idPrv *prv.Key
	if idPrv, e = prv.GenerateKey(); check(e) {
		return
	}
	id := nonce.NewID()
	introducerHops := client.SessionManager.GetSessionsAtHop(2)
	var introducer *SessionData
	if len(introducerHops) > 1 {
		cryptorand.Shuffle(len(introducerHops), func(i, j int) {
			introducerHops[i], introducerHops[j] = introducerHops[j], introducerHops[i]
		})
	}
	introducer = introducerHops[0]
	returnHops := client.SessionManager.GetSessionsAtHop(5)
	var returner *SessionData
	if len(returnHops) > 1 {
		cryptorand.Shuffle(len(returnHops), func(i, j int) {
			returnHops[i], returnHops[j] = returnHops[j],
				returnHops[i]
		})
	}
	returner = returnHops[0]
	svc := &Service{
		Port:      2345,
		RelayRate: 43523,
		Transport: NewSim(64),
	}
	client.SendHiddenService(id, idPrv, time.Now().Add(time.Hour), returner,
		introducer, svc, func(id nonce.ID, ifc interface{},
			b slice.Bytes) (e error) {
			log.I.S("hidden service callback", id, ifc, b.ToBytes())
			return
		})
	log2.SetLogLevel(log2.Trace)
	// Now query everyone for the intro.
	idPub := pub.Derive(idPrv)
	peers := clients[1:]
	log.D.Ln("client address", client.GetLocalNodeAddressString())
	for i := range peers {
		wg.Add(1)
		counter.Inc()
		log.T.Ln("peer", i)
		if len(returnHops) > 1 {
			cryptorand.Shuffle(len(returnHops), func(i, j int) {
				returnHops[i], returnHops[j] = returnHops[j], returnHops[i]
			})
		}
		if len(introducerHops) > 1 {
			cryptorand.Shuffle(len(introducerHops), func(i, j int) {
				introducerHops[i], introducerHops[j] = introducerHops[j], introducerHops[i]
			})
		}
		client.SendIntroQuery(id, idPub, introducerHops[0], returnHops[0],
			func(in *Intro) {
				wg.Done()
				counter.Dec()
				log.I.Ln("success",
					in.ID, in.Key.ToBase32Abbreviated(), in.AddrPort)
			})
		wg.Wait()
	}
	quit.Q()
}
