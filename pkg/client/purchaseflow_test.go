package client

import (
	"sync"
	"testing"

	"github.com/indra-labs/indra/pkg/key/prv"
	"github.com/indra-labs/indra/pkg/nonce"
)

func TestPurchaseFlow(t *testing.T) {
	// This tests the full process of establishing sessions and mocks
	// payment and then runs a test on the bandwidth accounting system.
	clients, e := CreateMockCircuitClients(6)
	if check(e) {
		t.Error(e)
		t.FailNow()
	}
	// Start up the clients.
	for _, v := range clients {
		go v.Start()
	}
	// quit := qu.T()
	// go func() {
	// 	time.Sleep(time.Second * 3)
	// 	quit.Q()
	// }()
	// <-quit.Wait()
	selected := clients[0].Nodes.Select(SimpleSelector, clients[1].Node, 4)
	// next to send out keys for the return hops
	returnNodes := selected[2:]
	var rtnHdr, rtnPld [2]*prv.Key
	var confirmation [2]nonce.ID
	var wait sync.WaitGroup
	for i := range returnNodes {
		wait.Add(1)
		confirmation[i], e = clients[0].
			SendKeys(returnNodes[i].ID, func(cf nonce.ID) {
				log.I.S("confirmed", cf)
				wait.Done()
			})
	}
	log.I.S(confirmation)
	log.I.S(rtnHdr, rtnPld)
	wait.Wait()
	// now to do the purchase

	for _, v := range clients {
		v.Shutdown()
	}
}
