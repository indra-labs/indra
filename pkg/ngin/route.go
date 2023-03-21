package ngin

import (
	"crypto/cipher"
	"net/netip"
	"reflect"
	"time"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/ciph"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/cloak"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/signer"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/ngin/magic"
	"git-indra.lan/indra-labs/indra/pkg/util/zip"
)

const (
	RouteMagic = "ro"
	RouteLen   = magic.Len + cloak.Len + pub.KeyLen + nonce.IVLen +
		zip.ReplyLen
)

func RoutePrototype() Onion { return &Route{} }

func init() { Register(RouteMagic, RoutePrototype) }

type Route struct {
	HiddenService *pub.Key
	HiddenCloaked cloak.PubKey
	Sender        *prv.Key
	SenderPub     *pub.Key
	nonce.IV
	// ------- the rest is encrypted to the HiddenService/Sender keys.
	*zip.Reply
	Onion
}

func (o Skins) Route(id nonce.ID, k *pub.Key, ks *signer.KeySet,
	ep *ExitPoint) Skins {
	
	oo := &Route{
		HiddenService: k,
		Sender:        ks.Next(),
		IV:            nonce.New(),
		Reply: &zip.Reply{
			ID:      id,
			Ciphers: GenCiphers(ep.Keys, ep.ReturnPubs),
			Nonces:  ep.Nonces,
		},
	}
	oo.SenderPub = pub.Derive(oo.Sender)
	oo.HiddenCloaked = cloak.GetCloak(k)
	return append(o, oo)
}

func (x *Route) Magic() string { return TmplMagic }

func (x *Route) Encode(s *zip.Splice) (e error) {
	iv := nonce.New()
	log.T.S("encoding", reflect.TypeOf(x),
		cloak.GetCloak(x.HiddenService), pub.Derive(x.Sender), iv,
		x.Reply,
	)
	s.Magic(RouteMagic).
		Cloak(x.HiddenService).
		Pubkey(pub.Derive(x.Sender)).
		IV(iv)
	start := s.GetCursor()
	s.Reply(x.Reply)
	var blk cipher.Block
	// Encrypt the message!
	if blk = ciph.GetBlock(x.Sender, x.HiddenService); check(e) {
		return
	}
	ciph.Encipher(blk, x.IV, s.GetRange(start, -1))
	return
}

func (x *Route) Decode(s *zip.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), RouteLen-magic.Len,
		RouteMagic); check(e) {
		return
	}
	s.ReadCloak(&x.HiddenCloaked).
		ReadPubkey(&x.SenderPub).
		ReadIV(&x.IV)
	return
}

// Decrypt decrypts the rest of a message after the Route segment if the
// recipient has the hidden service private key.
func (x *Route) Decrypt(prk *prv.Key, s *zip.Splice) {
	ciph.Encipher(ciph.GetBlock(prk, x.HiddenService), x.IV,
		s.GetCursorToEnd())
	// And now we can see the reply field for the return trip.
	if x.Reply == nil {
		x.Reply = &zip.Reply{}
	}
	s.ReadReply(x.Reply)
}

func (x *Route) Len() int { return RouteLen }

func (x *Route) Wrap(inner Onion) {}

func (x *Route) Handle(s *zip.Splice, p Onion, ng *Engine) (e error) {
	
	log.D.Ln(ng.GetLocalNodeAddressString(), "handling route",
		// ng.HiddenRouting.KnownIntros, ng.HiddenRouting.MyIntros,
	)
	hc := ng.FindCloakedHiddenService(x.HiddenCloaked)
	if hc == nil {
		log.T.Ln("no matching hidden service key found from cloaked key")
		return
	}
	x.HiddenService, e = pub.FromBytes((*hc)[:])
	log.D.Ln("route key", *hc)
	hcl := *hc
	if hh, ok := ng.HiddenRouting.HiddenServices[hcl]; ok {
		log.D.S("we are the hidden service", hh)
		// We have the keys to unwrap this one.
		x.Decrypt(hh.Prv, s)
		log.D.Ln(s)
		// ng.HandleMessage(s, x)
		return
	}
	// If we aren't the hidden service then we have maybe got the header to
	// open a connection from the hidden client to the hidden service.
	// The message is encrypted to them and will be recognised and accepted.
	var tryCount int
	for {
		log.I.Ln("trycount", tryCount)
		hb := ng.HiddenRouting.FindIntroduction(hcl)
		if hb != nil {
			log.D.S("found route", hb.ID, hb.AddrPort.String(),
				hb.Bytes.ToBytes())
			
			hops := []byte{3, 4, 5}
			ss := make(Sessions, len(hops))
			ng.SelectHops(hops, ss)
			for i := range ss {
				log.D.Ln(ss[i].Hop, ss[i].Node.AddrPort.String())
			}
			log.D.S("formulating reply...",
				s.GetRange(-1, s.GetCursor()).ToBytes(),
				s.GetRange(s.GetCursor(), -1).ToBytes(),
			)
			rb := FormatReply(hb.Bytes, s.GetRange(-1, -1),
				hb.Ciphers, hb.Nonces)
			log.D.S(rb.GetRange(-1, -1).ToBytes())
			ng.HandleMessage(rb, x)
			
			// We have to get another one before we can do this again.
			ng.HiddenRouting.Delete(hcl)
			log.D.Ln("deleted", hb.Intro.Key.ToBase32Abbreviated())
			return
		}
		// We have to retry a few times before giving up if the intro isn't
		// found.
		tryCount++
		if tryCount > 2 {
			log.D.Ln("finished handling route")
			log.D.S("HiddenRouting", ng.HiddenRouting.KnownIntros,
				ng.HiddenRouting.MyIntros, ng.HiddenRouting.HiddenServices)
			return
		}
		select {
		case <-time.After(time.Second): // * time.Duration(tryCount*tryCount)):
			continue
		case <-ng.C.Wait():
			return
		}
	}
}

func MakeRoute(id nonce.ID, k *pub.Key, ks *signer.KeySet,
	alice, bob *SessionData, c Circuit) Skins {
	
	headers := GetHeaders(alice, bob, c, ks)
	return Skins{}.
		RoutingHeader(headers.Forward).
		Route(id, k, ks, headers.ExitPoint()).
		RoutingHeader(headers.Return)
}

func (ng *Engine) SendRoute(k *pub.Key, ap *netip.AddrPort,
	hook Callback) {
	
	ng.FindNodeByAddrPort(ap)
	var ss *SessionData
	ng.IterateSessions(func(s *SessionData) bool {
		if s.Node.AddrPort.String() == ap.String() {
			ss = s
			return true
		}
		return false
	})
	if ss == nil {
		log.E.Ln(ng.GetLocalNodeAddressString(),
			"could not find session for address", ap.String())
		return
	}
	log.D.Ln("sending route", k.ToBase32Abbreviated())
	hops := StandardCircuit()
	s := make(Sessions, len(hops))
	s[2] = ss
	se := ng.SelectHops(hops, s)
	var c Circuit
	copy(c[:], se)
	o := MakeRoute(nonce.NewID(), k, ng.KeySet, c[4], ss, c)
	log.D.Ln("doing accounting")
	res := ng.PostAcctOnion(o)
	log.D.Ln("sending out route request onion")
	ng.SendWithOneHook(c[0].Node.AddrPort, res, hook, ng.PendingResponses)
}
