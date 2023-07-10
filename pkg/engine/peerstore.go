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
			log.D.Ln("received message from gossip network")
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
	log.D.S("decoded", c)
	var ok bool
	switch c.(type) {
	case *addresses.Ad:
		log.D.Ln("received", reflect.TypeOf(c), "from gossip network")
		var addr *addresses.Ad
		if addr, ok = c.(*addresses.Ad); !ok {
			return fmt.Errorf(ErrWrongTypeDecode,
				addresses.Magic, reflect.TypeOf(c).String())
		} else if !addr.Validate() {
			return errors.New("addr ad failed validation")
		}
		// If we got to here now we can add to the PeerStore.
		log.D.S("new ad for address:", c)
		var id peer.ID
		if id, e = peer.IDFromPublicKey(addr.Key); fails(e) {
			return
		}
		if e = ng.Listener.Host.
			Peerstore().Put(id, addresses.Magic, s.GetAll().ToBytes()); fails(e) {
			return
		}
	case *intro.Ad:
		log.D.Ln("received", reflect.TypeOf(c), "from gossip network")
		var intr *intro.Ad
		if intr, ok = c.(*intro.Ad); !ok {
			return fmt.Errorf(ErrWrongTypeDecode,
				intro.Magic, reflect.TypeOf(c).String())
		} else if !intr.Validate() {
			return errors.New("intro ad failed validation")
		}
		// If we got to here now we can add to the PeerStore.
		log.D.S("new ad for intro:", c)
		var id peer.ID
		if id, e = peer.IDFromPublicKey(intr.Key); fails(e) {
			return
		}
		if e = ng.Listener.Host.
			Peerstore().Put(id, intro.Magic, s.GetAll().ToBytes()); fails(e) {
			return
		}
	case *load.Ad:
		log.D.Ln("received", reflect.TypeOf(c), "from gossip network")
		var lod *load.Ad
		if lod, ok = c.(*load.Ad); !ok {
			return fmt.Errorf(ErrWrongTypeDecode,
				addresses.Magic, reflect.TypeOf(c).String())
		} else if !lod.Validate() {
			return errors.New("load ad failed validation")
		}
		// If we got to here now we can add to the PeerStore.
		log.D.S("new ad for load:", c)
		var id peer.ID
		if id, e = peer.IDFromPublicKey(lod.Key); fails(e) {
			return
		}
		if e = ng.Listener.Host.
			Peerstore().Put(id, services.Magic, s.GetAll().ToBytes()); fails(e) {
			return
		}
	case *peer2.Ad:
		log.D.Ln("received", reflect.TypeOf(c), "from gossip network")
		var pa *peer2.Ad
		if pa, ok = c.(*peer2.Ad); !ok {
			return fmt.Errorf(ErrWrongTypeDecode,
				peer2.Magic, reflect.TypeOf(c).String())
		} else if !pa.Validate() {
			return errors.New("peer ad failed validation")
		}
		// If we got to here now we can add to the PeerStore.
		log.D.S("new ad for peer:", c)
		var id peer.ID
		if id, e = peer.IDFromPublicKey(pa.Key); fails(e) {
			return
		}
		if e = ng.Listener.Host.
			Peerstore().Put(id, peer2.Magic, s.GetAll().ToBytes()); fails(e) {
			return
		}
	case *services.Ad:
		log.D.Ln("received", reflect.TypeOf(c), "from gossip network")
		log.D.S("message", c)
		var sa *services.Ad
		if sa, ok = c.(*services.Ad); !ok {
			return fmt.Errorf(ErrWrongTypeDecode,
				services.Magic, reflect.TypeOf(c).String())
		} else if !sa.Validate() {
			return errors.New("services ad failed validation")
		}
		// If we got to here now we can add to the PeerStore.
		log.D.S("new ad for services:", c)
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
