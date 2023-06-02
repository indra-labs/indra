//go:build failingtests

package headers

import (
	"context"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/engine"
	"github.com/indra-labs/indra/pkg/engine/sessions"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/cryptorand"
	"github.com/indra-labs/indra/pkg/util/qu"
	"github.com/indra-labs/indra/pkg/util/slice"
	"sync"
	"testing"
	"time"
)

func TestClient_SendGetBalance(t *testing.T) {
	log2.SetLogLevel(log2.Debug)
	var clients []*engine.Engine
	var e error
	ctx, cancel := context.WithCancel(context.Background())
	if clients, e = engine.CreateNMockCircuitsWithSessions(2, 2,
		ctx); engine.fails(e) {
		t.Error(e)
		t.FailNow()
	}
	client := clients[0]
	engine.log.D.Ln("client", client.Manager.GetLocalNodeAddressString())
	// Start up the clients.
	for _, v := range clients {
		go v.Start()
	}
	quit := qu.T()
	var wg sync.WaitGroup
	go func() {
		select {
		case <-time.After(5 * time.Second):
		case <-quit:
			return
		}
		cancel()
		quit.Q()
		t.Error("SendGetBalance test failed")
	}()
out:
	for i := 1; i < len(clients[0].Manager.Sessions)-1; i++ {
		wg.Add(1)
		returnHops := client.Manager.GetSessionsAtHop(5)
		var returner *sessions.Data
		if len(returnHops) > 1 {
			cryptorand.Shuffle(len(returnHops), func(i, j int) {
				returnHops[i], returnHops[j] = returnHops[j],
					returnHops[i]
			})
		}
		returner = returnHops[0]
		clients[0].SendGetBalance(returner, clients[0].Manager.Sessions[i],
			func(cf nonce.ID, ifc interface{}, b slice.Bytes) (e error) {
				engine.log.I.Ln("success")
				wg.Done()
				quit.Q()
				return
			})
		select {
		case <-quit:
			break out
		default:
		}
		wg.Wait()
	}
	quit.Q()
	cancel()
	for _, v := range clients {
		v.Shutdown()
	}
}
