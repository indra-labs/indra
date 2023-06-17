package engine

import (
	"github.com/indra-labs/indra"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/onions/adaddress"
	"github.com/indra-labs/indra/pkg/onions/adintro"
	"github.com/indra-labs/indra/pkg/onions/adload"
	"github.com/indra-labs/indra/pkg/onions/adpeer"
	"github.com/indra-labs/indra/pkg/onions/adservices"
	"github.com/indra-labs/indra/pkg/util/splice"
	"os"
	"testing"
	"time"

	"github.com/indra-labs/indra/pkg/engine/transport"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
)

func TestEngine_PeerStore(t *testing.T) {
	if indra.CI == "false" {
		log2.SetLogLevel(log2.Debug)
	}
	const nTotal = 26
	var cancel func()
	var e error
	var engines []*Engine
	var seed string
	for i := 0; i < nTotal; i++ {
		dataPath, err := os.MkdirTemp(os.TempDir(), "badger")
		if err != nil {
			t.FailNow()
		}
		var eng *Engine
		if eng, cancel, e = CreateMockEngine(seed, dataPath); fails(e) {
			return
		}
		engines = append(engines, eng)
		if i == 0 {
			seed = transport.GetHostAddress(eng.Listener.Host)
		}
		defer os.RemoveAll(dataPath)
		go eng.Start()
	}
	time.Sleep(time.Second)
	newAddressAd := adaddress.New(nonce.NewID(),
		engines[0].Manager.GetLocalNodeIdentityPrv(),
		engines[0].Listener.Host.Addrs()[0],
		time.Now().Add(time.Hour*24*7))
	sa := splice.New(newAddressAd.Len())
	if e = newAddressAd.Encode(sa); fails(e) {
		t.FailNow()
	}
	if e = engines[0].SendAd(sa.GetAll()); fails(e) {
		t.FailNow()
	}
	time.Sleep(time.Second)
	newIntroAd := adintro.New(nonce.NewID(),
		engines[0].Manager.GetLocalNodeIdentityPrv(),
		engines[0].Manager.GetLocalNodeAddress(),
		20000, 443,
		time.Now().Add(time.Hour*24*7))
	si := splice.New(newIntroAd.Len())
	if e = newIntroAd.Encode(si); fails(e) {
		t.FailNow()
	}
	if e = engines[0].SendAd(si.GetAll()); fails(e) {
		t.FailNow()
	}
	time.Sleep(time.Second)
	newLoadAd := adload.New(nonce.NewID(),
		engines[0].Manager.GetLocalNodeIdentityPrv(),
		17,
		time.Now().Add(time.Hour*24*7))
	sl := splice.New(newLoadAd.Len())
	if e = newLoadAd.Encode(sl); fails(e) {
		t.FailNow()
	}
	if e = engines[0].SendAd(sl.GetAll()); fails(e) {
		t.FailNow()
	}
	time.Sleep(time.Second)
	newPeerAd := adpeer.New(nonce.NewID(),
		engines[0].Manager.GetLocalNodeIdentityPrv(),
		20000,
		time.Now().Add(time.Hour*24*7))
	log.D.S("peer ad", newPeerAd)
	sp := splice.New(newPeerAd.Len())
	if e = newPeerAd.Encode(sp); fails(e) {
		t.FailNow()
	}
	if e = engines[0].SendAd(sp.GetAll()); fails(e) {
		t.FailNow()
	}
	time.Sleep(time.Second * 1)
	newServiceAd := adservices.New(nonce.NewID(),
		engines[0].Manager.GetLocalNodeIdentityPrv(),
		[]adservices.Service{{20000, 54321}},
		time.Now().Add(time.Hour*24*7))
	ss := splice.New(newServiceAd.Len())
	if e = newServiceAd.Encode(ss); fails(e) {
		t.FailNow()
	}
	if e = engines[0].SendAd(ss.GetAll()); fails(e) {
		t.FailNow()
	}
	time.Sleep(time.Second)
	cancel()
	for i := range engines {
		engines[i].Shutdown()
	}
}
