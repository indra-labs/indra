package engine

import (
	"github.com/indra-labs/indra"
	"os"
	"testing"
	"time"

	"github.com/indra-labs/indra/pkg/engine/transport"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
)

func TestEngine_Dispatcher(t *testing.T) {
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
			seed = transport.GetHostAddress(eng.Manager.Listener.Host)
		}
		defer os.RemoveAll(dataPath)
		go eng.Start()
		log.D.Ln("started engine", i)
	}
	time.Sleep(time.Second * 1)
	cancel()
	for i := range engines {
		engines[i].Shutdown()
	}
}
