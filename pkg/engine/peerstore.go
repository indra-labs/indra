package engine

import (
	"context"
	"errors"
	"fmt"
	"github.com/indra-labs/indra/pkg/ad"
	"github.com/indra-labs/indra/pkg/onions/adload"
	"github.com/indra-labs/indra/pkg/util/slice"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"reflect"

	"github.com/indra-labs/indra/pkg/onions/adaddress"
	"github.com/indra-labs/indra/pkg/onions/adintro"
	"github.com/indra-labs/indra/pkg/onions/adpeer"
	"github.com/indra-labs/indra/pkg/onions/adservices"
	"github.com/indra-labs/indra/pkg/onions/reg"
	"github.com/indra-labs/indra/pkg/util/splice"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

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

func (ng *Engine) SendAd(a slice.Bytes) (e error) {
	return ng.topic.Publish(ng.ctx, a)
}

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

const ErrWrongTypeDecode = "magic '%s' but type is '%s'"

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
	case *adaddress.Ad:
		log.D.Ln("received", reflect.TypeOf(c), "from gossip network")
		var addr *adaddress.Ad
		if addr, ok = c.(*adaddress.Ad); !ok {
			return fmt.Errorf(ErrWrongTypeDecode,
				adaddress.Magic, reflect.TypeOf(c).String())
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
			Peerstore().Put(id, adaddress.Magic, s.GetAll().ToBytes()); fails(e) {
			return
		}
	case *adintro.Ad:
		log.D.Ln("received", reflect.TypeOf(c), "from gossip network")
		var intr *adintro.Ad
		if intr, ok = c.(*adintro.Ad); !ok {
			return fmt.Errorf(ErrWrongTypeDecode,
				adintro.Magic, reflect.TypeOf(c).String())
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
			Peerstore().Put(id, adintro.Magic, s.GetAll().ToBytes()); fails(e) {
			return
		}
	case *adload.Ad:
		log.D.Ln("received", reflect.TypeOf(c), "from gossip network")
		var lod *adload.Ad
		if lod, ok = c.(*adload.Ad); !ok {
			return fmt.Errorf(ErrWrongTypeDecode,
				adaddress.Magic, reflect.TypeOf(c).String())
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
			Peerstore().Put(id, adservices.Magic, s.GetAll().ToBytes()); fails(e) {
			return
		}
	case *adpeer.Ad:
		log.D.Ln("received", reflect.TypeOf(c), "from gossip network")
		var pa *adpeer.Ad
		if pa, ok = c.(*adpeer.Ad); !ok {
			return fmt.Errorf(ErrWrongTypeDecode,
				adpeer.Magic, reflect.TypeOf(c).String())
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
			Peerstore().Put(id, adpeer.Magic, s.GetAll().ToBytes()); fails(e) {
			return
		}
	case *adservices.Ad:
		log.D.Ln("received", reflect.TypeOf(c), "from gossip network")
		var sa *adservices.Ad
		if sa, ok = c.(*adservices.Ad); !ok {
			return fmt.Errorf(ErrWrongTypeDecode,
				adservices.Magic, reflect.TypeOf(c).String())
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
			Peerstore().Put(id, adservices.Magic, s.GetAll().ToBytes()); fails(e) {
			return
		}
	}
	return
}

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
