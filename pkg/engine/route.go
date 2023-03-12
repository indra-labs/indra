package engine

import (
	"time"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/engine/magic"
	"git-indra.lan/indra-labs/indra/pkg/util/octet"
)

const (
	RouteMagic = "ro"
	RouteLen   = magic.Len + nonce.IDLen + pub.KeyLen
)

func RoutePrototype() Onion { return &Route{} }

func init() { Register(RouteMagic, RoutePrototype) }

type Route struct {
	HiddenService, Receiver *pub.Key
}

func (o Skins) Route(key, receiver *pub.Key) Skins {
	return append(o, &Route{key, receiver})
}

func NewRoute(key, receiver *pub.Key) *Route {
	return &Route{key, receiver}
}

func (x *Route) Magic() string { return TmplMagic }

func (x *Route) Encode(s *octet.Splice) (e error) {
	s.Pubkey(x.HiddenService).Pubkey(x.Receiver)
	return
}

func (x *Route) Decode(s *octet.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), RouteLen-magic.Len,
		RouteMagic); check(e) {
		return
	}
	s.ReadPubkey(&x.HiddenService).ReadPubkey(&x.Receiver)
	return
}

func (x *Route) Len() int { return RouteLen }

func (x *Route) Wrap(inner Onion) {}

func (x *Route) Handle(s *octet.Splice, p Onion,
	ng *Engine) (e error) {
	
	// If we have an intro header we now send a request out to the hidden
	// service using the header we have cached.
	hsb := x.HiddenService.ToBytes()
	var tryCount int
	for {
		hb := ng.Introductions.Find(hsb)
		if hb != nil {
			// We have to get another one before we can do this again.
			ng.Introductions.Delete(hsb)
			if intro := ng.Introductions.FindKnownIntro(hsb); intro != nil {
				go func() {
					dest := intro.AddrPort
					
					_ = dest
					select {
					case <-time.After(time.Second * 5):
						return
					case <-ng.C.Wait():
						return
					}
				}()
				return
			}
		}
		// We have to retry a few times before giving up if the intro isn't
		// found.
		tryCount++
		if tryCount > 5 {
			return
		}
		select {
		case <-time.After(time.Second):
			continue
		case <-ng.C.Wait():
			return
		}
		
	}
}
