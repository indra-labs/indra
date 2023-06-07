package engine

import (
	"github.com/indra-labs/indra"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/onions/adpeer"
	"github.com/indra-labs/indra/pkg/util/splice"
	"os"
	"testing"
	"time"

	"github.com/indra-labs/indra/pkg/engine/transport"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
)

func TestEngine_PeerStore(t *testing.T) {
	if indra.CI == "false" {
		log2.SetLogLevel(log2.Trace)
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
	newAd := adpeer.New(nonce.NewID(),
		engines[0].Manager.GetLocalNodeIdentityPrv(), 20000)
	s := splice.New(newAd.Len())
	if e = newAd.Encode(s); fails(e) {
		t.FailNow()
	}

	time.Sleep(time.Second * 3)
	cancel()
	for i := range engines {
		engines[i].Shutdown()
	}
}
