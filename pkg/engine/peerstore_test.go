package engine

import (
	"context"
	"github.com/indra-labs/indra"
	"testing"
	"time"

	log2 "github.com/indra-labs/indra/pkg/proc/log"
)

func pauza() {
	time.Sleep(time.Second)
}

//
// func TestEngine_PeerStore(t *testing.T) {
// 	if indra.CI == "false" {
// 		log2.SetLogLevel(log2.Trace)
// 	}
// 	const nTotal = 10
// 	var e error
// 	var engines []*Engine
// 	var cleanup func()
// 	ctx, _ := context.WithCancel(context.Background())
// 	engines, cleanup, e = CreateAndStartMockEngines(nTotal, ctx)
// 	adz := engines[0].Listener.Host.Addrs()
// 	addrs := make([]*netip.AddrPort, len(adz))
// 	for i := range adz {
// 		addy, _ := multi.AddrToAddrPort(adz[i])
// 		addrs[i] = &addy
// 	}
// 	// To ensure every peer will get the gossip:
// 	pauza()
// 	newAddressAd := addresses.New(nonce.NewID(),
// 		engines[0].Mgr().GetLocalNodeIdentityPrv(),
// 		addrs,
// 		time.Now().Add(time.Hour*24*7))
// 	sa := splice.New(newAddressAd.Len())
// 	if e = newAddressAd.Encode(sa); fails(e) {
// 		t.FailNow()
// 	}
// 	if e = engines[0].SendAd(sa.GetAll()); fails(e) {
// 		t.FailNow()
// 	}
// 	newIntroAd := intro.New(nonce.NewID(),
// 		engines[0].Mgr().GetLocalNodeIdentityPrv(),
// 		engines[0].Mgr().GetLocalNode().Identity.Pub,
// 		20000, 443,
// 		time.Now().Add(time.Hour*24*7))
// 	si := splice.New(newIntroAd.Len())
// 	if e = newIntroAd.Encode(si); fails(e) {
// 		t.FailNow()
// 	}
// 	if e = engines[0].SendAd(si.GetAll()); fails(e) {
// 		t.FailNow()
// 	}
// 	newLoadAd := load.New(nonce.NewID(),
// 		engines[0].Mgr().GetLocalNodeIdentityPrv(),
// 		17,
// 		time.Now().Add(time.Hour*24*7))
// 	sl := splice.New(newLoadAd.Len())
// 	if e = newLoadAd.Encode(sl); fails(e) {
// 		t.FailNow()
// 	}
// 	if e = engines[0].SendAd(sl.GetAll()); fails(e) {
// 		t.FailNow()
// 	}
// 	newPeerAd := peer.New(nonce.NewID(),
// 		engines[0].Mgr().GetLocalNodeIdentityPrv(),
// 		20000,
// 		time.Now().Add(time.Hour*24*7))
// 	sp := splice.New(newPeerAd.Len())
// 	if e = newPeerAd.Encode(sp); fails(e) {
// 		t.FailNow()
// 	}
// 	if e = engines[0].SendAd(sp.GetAll()); fails(e) {
// 		t.FailNow()
// 	}
// 	newServiceAd := services.New(nonce.NewID(),
// 		engines[0].Mgr().GetLocalNodeIdentityPrv(),
// 		[]services.Service{{20000, 54321}, {10000, 42221}},
// 		time.Now().Add(time.Hour*24*7))
// 	ss := splice.New(newServiceAd.Len())
// 	if e = newServiceAd.Encode(ss); fails(e) {
// 		t.FailNow()
// 	}
// 	if e = engines[0].SendAd(ss.GetAll()); fails(e) {
// 		t.FailNow()
// 	}
// 	pauza()
// 	cleanup()
// }

func TestEngine_PeerStoreDiscovery(t *testing.T) {
	if indra.CI == "false" {
		log2.SetLogLevel(log2.Trace)
	}
	const nTotal = 20
	var e error
	var engines []*Engine
	var cleanup func()
	ctx, _ := context.WithCancel(context.Background())
	if engines, cleanup, e = CreateAndStartMockEngines(nTotal, ctx); fails(e) {
		t.FailNow()
	}
	_ = engines
	time.Sleep(time.Second * 8)
	cleanup()
	pauza()
}
