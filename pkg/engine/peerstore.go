package engine

import (
	"context"
	"errors"
	"fmt"
	"github.com/indra-labs/indra/pkg/onions/adload"
	"github.com/libp2p/go-libp2p/core/peer"
	"reflect"

	"github.com/indra-labs/indra/pkg/ad"
	"github.com/indra-labs/indra/pkg/onions/adaddress"
	"github.com/indra-labs/indra/pkg/onions/adintro"
	"github.com/indra-labs/indra/pkg/onions/adpeer"
	"github.com/indra-labs/indra/pkg/onions/adservices"
	"github.com/indra-labs/indra/pkg/onions/ont"
	"github.com/indra-labs/indra/pkg/onions/reg"
	"github.com/indra-labs/indra/pkg/util/splice"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

func (ng *Engine) SendAd(a ad.Ad) (e error) {
	return ng.topic.Publish(ng.ctx, ont.Encode(a).GetAll())
}

func (ng *Engine) RunAdHandler(handler func(p *pubsub.Message,
	ctx context.Context) (e error)) {

	// Since the frequency of updates should be around 1 hour we run here only
	// one thread here. Relays indicate their loading as part of the response
	// message protocol for ranking in the session cache.
	go func(ng *Engine) {
	out:
		for {
			var m *pubsub.Message
			var e error
			log.D.Ln("waiting for next message from gossip network")
			if m, e = ng.sub.Next(ng.ctx); e != nil {
				continue
			}
			log.D.Ln("received message from gossip network")
			if e = handler(m, ng.ctx); fails(e) {
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

func (ng *Engine) HandleAd(p *pubsub.Message, ctx context.Context) (e error) {
	if len(p.Data) < 1 {
		log.E.Ln("received slice of no length")
		return
	}
	s := splice.NewFrom(p.Data)
	c := reg.Recognise(s)
	if c == nil {
		return errors.New("message not recognised")
	}
	if e = c.Decode(s); fails(e) {
		return
	}
	switch c.(type) {
	case *adaddress.Ad:
		log.D.Ln("received", reflect.TypeOf(c), "from gossip network")
		if addr, ok := c.(*adaddress.Ad); !ok {
			return fmt.Errorf(ErrWrongTypeDecode,
				adaddress.Magic, reflect.TypeOf(c).String())
		} else {
			// If we got to here now we can add to the PeerStore.
			log.D.S("new ad for services:", c)
			var id peer.ID
			if id, e = peer.IDFromPublicKey(addr.Key); fails(e) {
				return
			}
			if e = ng.Listener.Host.
				Peerstore().Put(id, "services", s.GetAll().ToBytes()); fails(e) {
				return
			}
		}
	case *adintro.Ad:
		log.D.Ln("received", reflect.TypeOf(c), "from gossip network")
		if intr, ok := c.(*adintro.Ad); !ok {
			return fmt.Errorf(ErrWrongTypeDecode,
				adintro.Magic, reflect.TypeOf(c).String())
		} else {
			// If we got to here now we can add to the PeerStore.
			log.D.S("new ad for services:", c)
			var id peer.ID
			if id, e = peer.IDFromPublicKey(intr.Key); fails(e) {
				return
			}
			if e = ng.Listener.Host.
				Peerstore().Put(id, "services", s.GetAll().ToBytes()); fails(e) {
				return
			}
		}
	case *adload.Ad:
		log.D.Ln("received", reflect.TypeOf(c), "from gossip network")
		if lod, ok := c.(*adload.Ad); !ok {
			return fmt.Errorf(ErrWrongTypeDecode,
				adaddress.Magic, reflect.TypeOf(c).String())
		} else {
			// If we got to here now we can add to the PeerStore.
			log.D.S("new ad for services:", c)
			var id peer.ID
			if id, e = peer.IDFromPublicKey(lod.Key); fails(e) {
				return
			}
			if e = ng.Listener.Host.
				Peerstore().Put(id, "services", s.GetAll().ToBytes()); fails(e) {
				return
			}
		}
	case *adservices.Ad:
		log.D.Ln("received", reflect.TypeOf(c), "from gossip network")
		if serv, ok := c.(*adservices.Ad); !ok {
			return fmt.Errorf(ErrWrongTypeDecode,
				adservices.Magic, reflect.TypeOf(c).String())
		} else {
			// If we got to here now we can add to the PeerStore.
			log.D.S("new ad for services:", c)
			var id peer.ID
			if id, e = peer.IDFromPublicKey(serv.Key); fails(e) {
				return
			}
			if e = ng.Listener.Host.
				Peerstore().Put(id, "services", s.GetAll().ToBytes()); fails(e) {
				return
			}
		}
	case *adpeer.Ad:
		log.D.Ln("received", reflect.TypeOf(c), "from gossip network")
		var ok bool
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
			Peerstore().Put(id, "peer", s.GetAll().ToBytes()); fails(e) {
			return
		}
	}
	return
}
