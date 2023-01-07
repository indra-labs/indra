package client

import (
	"testing"
	"time"

	"github.com/cybriq/qu"
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
	quit := qu.T()
	go func() {
		time.Sleep(time.Second * 3)
		quit.Q()
	}()
	selected := clients[0].Nodes.Select(SimpleSelector, clients[1].Node, 4)
	_ = selected
	<-quit.Wait()
	for _, v := range clients {
		v.Shutdown()
	}
}
