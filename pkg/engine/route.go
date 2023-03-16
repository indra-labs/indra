package engine

import (
	"time"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/signer"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/engine/magic"
	"git-indra.lan/indra-labs/indra/pkg/util/octet"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

const (
	RouteMagic = "ro"
	RouteLen   = magic.Len + nonce.IDLen + pub.KeyLen + ReverseHeaderLen
)

func RoutePrototype() Onion { return &Route{} }

func init() { Register(RouteMagic, RoutePrototype) }

type Route struct {
	HiddenService *pub.Key
	// Header is the 3 layer header to use with the following cipher and
	// nonces to package the return message.
	Header slice.Bytes
	// Ciphers is a set of 3 symmetric ciphers that are to be used in their
	// given order over the reply message from the service.
	Ciphers [3]sha256.Hash
	// Nonces are the nonces to use with the cipher when creating the
	// encryption for the reply message.
	Nonces [3]nonce.IV
	Onion
}

func (o Skins) Route(key *pub.Key, header slice.Bytes,
	point *ExitPoint) Skins {
	return append(o, &Route{
		HiddenService: key,
		Header:        header,
		Ciphers:       GenCiphers(point.Keys, point.ReturnPubs),
		Nonces:        point.Nonces,
		Onion:         NewTmpl(),
	})
}

func (x *Route) Magic() string { return TmplMagic }

func (x *Route) Encode(s *octet.Splice) (e error) {
	s.Magic(RouteMagic).
		Pubkey(x.HiddenService).
		Bytes(x.Header).
		HashTriple(x.Ciphers).
		IVTriple(x.Nonces)
	if x.Onion != nil {
		return x.Onion.Encode(s)
	}
	return
}

func (x *Route) Decode(s *octet.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), RouteLen-magic.Len,
		RouteMagic); check(e) {
		return
	}
	s.ReadPubkey(&x.HiddenService).
		ReadBytes(&x.Header).
		ReadHashTriple(&x.Ciphers).
		ReadIVTriple(&x.Nonces)
	return
}

func (x *Route) Len() int { return RouteLen + x.Onion.Len() }

func (x *Route) Wrap(inner Onion) {}

func (x *Route) Handle(s *octet.Splice, p Onion,
	ng *Engine) (e error) {
	
	log.D.Ln("handling route")
	// If we have an intro header we now send a request out to the hidden
	// service using the header we have cached.
	hsb := x.HiddenService.ToBytes()
	var tryCount int
	for {
		hb := ng.Introductions.Find(hsb)
		if hb != nil {
			log.D.S("found route",
				hb.Intro.ID,
				hb.Intro.Key.ToBase32Abbreviated(),
				hb.Intro.AddrPort.String(),
				hb.Intro.Expiry,
				hb.Intro.Sig,
				hb.Intro.Validate(),
				hb.Bytes.ToBytes(),
				hb.Ciphers,
				hb.Nonces,
			)
			// We have to get another one before we can do this again.
			ng.Introductions.Delete(hsb)
			log.D.Ln("deleted", hb.Intro.Key.ToBase32Abbreviated())
			hops := []byte{3, 4, 5}
			ss := make(Sessions, len(hops))
			ng.SelectHops(hops, ss)
			log.D.Ln("selected sessions", ss)
			n := GenNonces(3)
			r := &Routing{
				Sessions: [3]*SessionData{ss[0], ss[1], ss[2]},
				Keys: [3]*prv.Key{
					ss[0].HeaderPrv, ss[1].HeaderPrv, ss[2].HeaderPrv,
				},
				Nonces: [3]nonce.IV{n[0], n[1], n[2]},
			}
			hr := MakeHiddenRoute(hb.Intro.Key, hb.Bytes, r)
			ob := hr.Assemble()
			encoded := Encode(ob)
			rb := FormatReply(hb.Bytes.ToBytes(), encoded.GetRange(-1, -1),
				hb.Ciphers, hb.Nonces)
			log.D.S("hidden route reply", rb)
			ng.HandleMessage(rb, x)
			return
		}
		// We have to retry a few times before giving up if the intro isn't
		// found.
		tryCount++
		if tryCount > 5 {
			return
		}
		select {
		case <-time.After(time.Second * time.Duration(tryCount*tryCount)):
			continue
		case <-ng.C.Wait():
			return
		}
	}
}

func MakeRoute(hs *pub.Key, header slice.Bytes, target *SessionData, s Circuit,
	ks *signer.KeySet) Skins {
	headers := GetHeaders(target, s, ks)
	return Skins{}.
		RoutingHeader(headers.Forward).
		Route(hs, header, headers.ExitPoint()).
		RoutingHeader(headers.Return)
}

func (ng *Engine) SendRoute(hs *pub.Key, header slice.Bytes,
	target *SessionData, hook Callback) {
	
	log.D.Ln("sending route", hs.ToBase32Abbreviated())
	hops := StandardCircuit()
	s := make(Sessions, len(hops))
	s[2] = target
	se := ng.SelectHops(hops, s)
	var c Circuit
	copy(c[:], se)
	o := MakeRoute(hs, header, c[2], c, ng.KeySet)
	log.D.Ln("sending out route request onion")
	res := ng.PostAcctOnion(o)
	ng.SendWithOneHook(c[0].AddrPort, res, hook, ng.PendingResponses)
}
