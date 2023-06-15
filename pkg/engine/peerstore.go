package engine

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/indra-labs/indra/pkg/ad"
	"github.com/indra-labs/indra/pkg/onions/adaddress"
	"github.com/indra-labs/indra/pkg/onions/adintro"
	"github.com/indra-labs/indra/pkg/onions/adpeer"
	"github.com/indra-labs/indra/pkg/onions/adservice"
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
			if m, e = ng.sub.Next(ng.ctx); e != nil {
				continue
			}

			if e = handler(m, ng.ctx); fails(e) {
				continue
			}
			select {
			case <-ng.ctx.Done():
				break out
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
	switch c.Magic() {
	case adaddress.Magic:
		log.D.Ln("received", reflect.TypeOf(c), "from gossip network")
		if addr, ok := c.(*adaddress.Ad); !ok {
			return fmt.Errorf(ErrWrongTypeDecode,
				adaddress.Magic, reflect.TypeOf(c).String())
		} else {
			_ = addr
		}
	case adintro.Magic:
		log.D.Ln("received", reflect.TypeOf(c), "from gossip network")
		if intr, ok := c.(*adintro.Ad); !ok {
			return fmt.Errorf(ErrWrongTypeDecode,
				adintro.Magic, reflect.TypeOf(c).String())
		} else {
			_ = intr

		}
	case adservice.Magic:
		log.D.Ln("received", reflect.TypeOf(c), "from gossip network")
		if serv, ok := c.(*adservice.Ad); !ok {
			return fmt.Errorf(ErrWrongTypeDecode,
				adservice.Magic, reflect.TypeOf(c).String())
		} else {
			_ = serv

		}
	case adpeer.Magic:
		log.D.Ln("received", reflect.TypeOf(c), "from gossip network")
		if peer, ok := c.(*adpeer.Ad); !ok {
			return fmt.Errorf(ErrWrongTypeDecode,
				adpeer.Magic, reflect.TypeOf(c).String())
		} else if !peer.Validate() {
			return errors.New("peer ad failed validation")
		}
		// If we got to here now we can add to the PeerStore.

	}
	return
}
