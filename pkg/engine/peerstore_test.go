package engine

import (
	"context"
	"github.com/dgraph-io/badger/v3"
	"github.com/indra-labs/indra"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"strings"
	"testing"
	"time"
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
		// log2.SetLogLevel(log2.Trace)
	}
	const nTotal = 10
	var (
		e       error
		engines []*Engine
		cleanup func()
	)
	ctx, cancel := context.WithCancel(context.Background())
	if engines, cleanup, e = CreateAndStartMockEngines(nTotal, ctx,
		cancel); fails(e) {

		t.FailNow()
	}
	time.Sleep(time.Second * 3)
	// Send them all again after a bit to make sure everyone gets them.
	for i := range engines {
		if e = engines[i].SendAds(); fails(e) {
			t.FailNow()
		}
	}
	time.Sleep(time.Second * 3)
	if indra.CI == "false" {
		log2.SetLogLevel(log2.Debug)
	}
	var ec int
	entryCount := &ec
	for _, v := range engines {
		// check that all peers now have nTotal-1 distinct peer ads (of all 4
		// types)
		e = v.PeerstoreView(func(txn *badger.Txn) error {
			defer txn.Discard()
			opts := badger.DefaultIteratorOptions
			it := txn.NewIterator(opts)
			defer it.Close()
			var val []byte
			var adCount int
			for it.Rewind(); it.Valid(); it.Next() {
				k := string(it.Item().Key())
				if !strings.HasSuffix(k, "ad") {
					continue
				}
				val, e = it.Item().ValueCopy(nil)
				log.T.S(v.LogEntry("item "+k), val)
				adCount++
			}
			log.T.Ln("adCount", adCount)
			if adCount == (nTotal-1)*4 {
				*entryCount++
			}
			return nil
		})
	}
	if *entryCount != nTotal {
		t.Log("nodes did not gossip completely to each other, only",
			*entryCount, "nodes ad sets counted, not the expected",
			nTotal)
		if indra.CI == "false" {
			t.FailNow()
		}
	}
	cleanup()
	pauza()
}
