package engine

import (
	"github.com/indra-labs/indra"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
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
	prvKey := engines[0].Manager.GetLocalNodeIdentityPrv()
	newPeerAd := adpeer.New(nonce.NewID(),
		prvKey, 20000,
		time.Now().Add(time.Hour*24*7))
	if e = newPeerAd.Sign(prvKey); fails(e) {
		return
	}
	log.D.S("peer ad", newPeerAd)
	s := splice.New(newPeerAd.Len())
	if e = newPeerAd.Encode(s); fails(e) {
		t.FailNow()
	}
	if e = engines[0].SendAd(newPeerAd); fails(e) {
		t.FailNow()
	}
	time.Sleep(time.Second * 3)
	newServiceAd := adservices.New(nonce.NewID(),
		engines[0].Manager.GetLocalNodeIdentityPrv(),
		[]adservices.Service{{20000, 54321}},
		time.Now().Add(time.Hour*24*7))
	ss := splice.New(newServiceAd.Len())
	if e = newServiceAd.Encode(ss); fails(e) {
		t.FailNow()
	}
	if e = engines[0].SendAd(newServiceAd); fails(e) {
		t.FailNow()
	}
	time.Sleep(time.Second * 3)
	cancel()
	for i := range engines {
		engines[i].Shutdown()
	}
}
