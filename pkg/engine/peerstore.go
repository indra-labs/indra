package engine

import (
	"context"
	"errors"
	"fmt"
	"github.com/indra-labs/indra/pkg/ad"
	"github.com/indra-labs/indra/pkg/codec/ad/load"
	"github.com/indra-labs/indra/pkg/util/slice"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"reflect"
	"time"

	"github.com/indra-labs/indra/pkg/codec/ad/addresses"
	"github.com/indra-labs/indra/pkg/codec/ad/intro"
	peer2 "github.com/indra-labs/indra/pkg/codec/ad/peer"
	"github.com/indra-labs/indra/pkg/codec/ad/services"
	"github.com/indra-labs/indra/pkg/codec/reg"
	"github.com/indra-labs/indra/pkg/util/splice"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

// SetupGossip establishes a connection of a Host to the pubsub gossip network
// used by Indra to propagate peer metadata.
func SetupGossip(ctx context.Context, host host.Host,
	cancel func()) (PubSub *pubsub.PubSub, topic *pubsub.Topic,
	sub *pubsub.Subscription, e error) {

	if PubSub, e = pubsub.NewGossipSub(ctx, host); fails(e) {
		cancel()
		return
	}
	if topic, e = PubSub.Join(PubSubTopic); fails(e) {
		cancel()
		return
	}
	if sub, e = topic.Subscribe(); fails(e) {
		cancel()
		return
	}
	log.T.Ln("subscribed to", PubSubTopic, "topic on gossip network")
	return
}

// SendAd dispatches an encoded byte slice ostensibly of a peer advertisement to gossip to the rest of the network.
func (ng *Engine) SendAd(a slice.Bytes) (e error) {
	return ng.topic.Publish(ng.ctx, a)
}

// SendAds dispatches all ads in NodeAds. Primarily called at startup.
func (ng *Engine) SendAds() (e error) {
	na := ng.NodeAds
	ads := []ad.Ad{na.Address, na.Load, na.Peer, na.Services}
	for i := range ads {
		s := splice.New(ads[i].Len())
		ads[i].Encode(s)
		if e = ng.topic.Publish(ng.ctx, s.GetAll()); fails(e) {
			return
		}
	}
	return
}

// RunAdHandler listens to the gossip and dispatches messages to be handled and
// update the peerstore.
func (ng *Engine) RunAdHandler(handler func(p *pubsub.Message) (e error)) {

	// Since the frequency of updates should be around 1 hour we run here only
	// one thread here. Relays indicate their loading as part of the response
	// message protocol for ranking in the session cache.
	go func(ng *Engine) {
	out:
		for {
			var m *pubsub.Message
			var e error
			if m, e = ng.sub.Next(ng.ctx); e != nil {
				continue
			}
			if e = handler(m); fails(e) {
				continue
			}
			select {
			case <-ng.ctx.Done():
				log.D.Ln("shutting down ad handler")
				break out
			default:
			}
		}
		return
	}(ng)

	go func(ng *Engine) {
		log.D.Ln(ng.Listener.Host.Addrs(), "checking and updating peer information ads")
		// First time we want to do the thing straight away and update the peers
		// with a new ads.NodeAds.
		ng.gossip(time.NewTicker(time.Second))
		// Then after this we check once a second
	}(ng)
}

func (ng *Engine) gossip(tick *time.Ticker) {
	now := time.Now()
	first := true
out:
	for {
		if first {
			first = false
			// Send out all ads because we are starting up.
			ng.SendAds()
			// As all ads are sent we can return to the head of the loop.
			continue
		}
		// Check for already generated NodeAds, and make them first time if
		// needed.
		na := ng.NodeAds
		log.D.Ln(ng.Listener.Host.Addrs(), "gossip tick")
		switch {
		case na.Address == nil:
			log.D.Ln("updating peer address")

			fallthrough

		case na.Load == nil:
			log.D.Ln("updating peer load")

			fallthrough

		case na.Peer == nil:
			log.D.Ln("updating peer ad")

			fallthrough

		case na.Services == nil &&
			// But only if we have any services:
			len(ng.Mgr().GetLocalNode().Services) > 0:
			log.D.Ln("updating services")

			fallthrough
			// Next, check each entry has not expired:

		case na.Address.Expiry.Before(now):
			log.D.Ln("updating expired peer address")

			fallthrough

		case na.Load.Expiry.Before(now):
			log.D.Ln("updating expired load ad")

			fallthrough

		case na.Peer.Expiry.Before(now):
			log.D.Ln("updating peer ad")

			fallthrough

		case na.Services.Expiry.Before(now):
			log.D.Ln("updating peer services")

		}
		// Then, lastly, check if the ad content has changed due to
		// reconfiguration or other reasons such as a more substantial amount of
		// load or drop in load, or changed IP addresses.
		if first {
			first = false
			// Send out all ads because we are starting up.

		}
		// After all that is done, check if we are shutting down, if so exit.
		select {
		case <-ng.ctx.Done():
			break out
		case <-tick.C:
		}
	}

}

// ErrWrongTypeDecode indicates a message has the wrong magic.
const ErrWrongTypeDecode = "magic '%s' but type is '%s'"

// HandleAd correctly recognises, validates, and stores the ads in the peerstore.
func (ng *Engine) HandleAd(p *pubsub.Message) (e error) {
	if len(p.Data) < 1 {
		log.E.Ln("received slice of no length")
		return
	}
	s := splice.NewFrom(p.Data)
	c := reg.Recognise(s)
	if c == nil {
		return errors.New("ad not recognised")
	}
	if e = c.Decode(s); fails(e) {
		return
	}
	var ok bool
	switch c.(type) {
	case *addresses.Ad:
		log.D.Ln(ng.Listener.Host.Addrs()[0].String()+" received", reflect.TypeOf(c), "from gossip network")
		var addr *addresses.Ad
		if addr, ok = c.(*addresses.Ad); !ok {
			return fmt.Errorf(ErrWrongTypeDecode,
				addresses.Magic, reflect.TypeOf(c).String())
		} else if !addr.Validate() {
			return errors.New("addr ad failed validation")
		}
		// If we got to here now we can add to the PeerStore.
		var id peer.ID
		if id, e = peer.IDFromPublicKey(addr.Key); fails(e) {
			return
		}
		if e = ng.Listener.Host.
			Peerstore().Put(id, addresses.Magic, s.GetAll().ToBytes()); fails(e) {
			return
		}
	case *intro.Ad:
		log.D.Ln(ng.Listener.Host.Addrs()[0].String()+" received", reflect.TypeOf(c), "from gossip network")
		var intr *intro.Ad
		if intr, ok = c.(*intro.Ad); !ok {
			return fmt.Errorf(ErrWrongTypeDecode,
				intro.Magic, reflect.TypeOf(c).String())
		} else if !intr.Validate() {
			return errors.New("intro ad failed validation")
		}
		// If we got to here now we can add to the PeerStore.
		var id peer.ID
		if id, e = peer.IDFromPublicKey(intr.Key); fails(e) {
			return
		}
		if e = ng.Listener.Host.
			Peerstore().Put(id, intro.Magic, s.GetAll().ToBytes()); fails(e) {
			return
		}
	case *load.Ad:
		log.D.Ln(ng.Listener.Host.Addrs()[0].String()+" received", reflect.TypeOf(c), "from gossip network")
		var lod *load.Ad
		if lod, ok = c.(*load.Ad); !ok {
			return fmt.Errorf(ErrWrongTypeDecode,
				addresses.Magic, reflect.TypeOf(c).String())
		} else if !lod.Validate() {
			return errors.New("load ad failed validation")
		}
		// If we got to here now we can add to the PeerStore.
		var id peer.ID
		if id, e = peer.IDFromPublicKey(lod.Key); fails(e) {
			return
		}
		if e = ng.Listener.Host.
			Peerstore().Put(id, services.Magic, s.GetAll().ToBytes()); fails(e) {
			return
		}
	case *peer2.Ad:
		log.D.Ln(ng.Listener.Host.Addrs()[0].String()+" received", reflect.TypeOf(c), "from gossip network")
		var pa *peer2.Ad
		if pa, ok = c.(*peer2.Ad); !ok {
			return fmt.Errorf(ErrWrongTypeDecode,
				peer2.Magic, reflect.TypeOf(c).String())
		} else if !pa.Validate() {
			return errors.New("peer ad failed validation")
		}
		// If we got to here now we can add to the PeerStore.
		var id peer.ID
		if id, e = peer.IDFromPublicKey(pa.Key); fails(e) {
			return
		}
		if e = ng.Listener.Host.
			Peerstore().Put(id, peer2.Magic, s.GetAll().ToBytes()); fails(e) {
			return
		}
	case *services.Ad:
		log.D.Ln(ng.Listener.Host.Addrs()[0].String()+" received", reflect.TypeOf(c), "from gossip network")
		var sa *services.Ad
		if sa, ok = c.(*services.Ad); !ok {
			return fmt.Errorf(ErrWrongTypeDecode,
				services.Magic, reflect.TypeOf(c).String())
		} else if !sa.Validate() {
			return errors.New("services ad failed validation")
		}
		// If we got to here now we can add to the PeerStore.
		var id peer.ID
		if id, e = peer.IDFromPublicKey(sa.Key); fails(e) {
			return
		}
		if e = ng.Listener.Host.
			Peerstore().Put(id, services.Magic, s.GetAll().ToBytes()); fails(e) {
			return
		}
	}
	return
}

// GetPeerRecord queries the peerstore for an ad from a given peer.ID and the ad
// type key. The ad type keys are the same as the Magic of each ad type, to be
// simple.
func (ng *Engine) GetPeerRecord(id peer.ID, key string) (add ad.Ad, e error) {
	var a interface{}
	if a, e = ng.Listener.Host.Peerstore().Get(id, key); fails(e) {
		return
	}
	var ok bool
	var adb slice.Bytes
	if adb, ok = a.(slice.Bytes); !ok {
		e = errors.New("peer record did not decode slice.Bytes")
		return
	}
	if len(adb) < 1 {
		e = fmt.Errorf("record for peer ID %v key %s has expired", id, key)
	}
	s := splice.NewFrom(adb)
	c := reg.Recognise(s)
	if c == nil {
		e = errors.New(key + " peer record not recognised")
		return
	}
	if e = c.Decode(s); fails(e) {
		return
	}
	if add, ok = c.(ad.Ad); !ok {
		e = errors.New(key + " peer record did not decode as Ad")
	}
	return
}

// ClearPeerRecord places an empty slice into a peer record by way of deleting it.
//
// todo: these should be purged from the peerstore in a GC pass.
func (ng *Engine) ClearPeerRecord(id peer.ID, key string) (e error) {
	if _, e = ng.Listener.Host.Peerstore().Get(id, key); fails(e) {
		return
	}
	if e = ng.Listener.Host.
		Peerstore().Put(id, key, []byte{}); fails(e) {
		return
	}
	return
}
