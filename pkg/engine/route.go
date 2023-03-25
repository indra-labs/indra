package engine

import (
	"crypto/cipher"
	"net/netip"
	"reflect"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/ciph"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/cloak"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/signer"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/engine/magic"
)

const (
	RouteMagic = "ro"
	RouteLen   = magic.Len + cloak.Len + pub.KeyLen + nonce.IVLen +
		nonce.IDLen + 3*sha256.Len + 3*nonce.IVLen
)

func RoutePrototype() Onion       { return &Route{} }
func init()                       { Register(RouteMagic, RoutePrototype) }
func (x *Route) Magic() string    { return EndMagic }
func (x *Route) Len() int         { return RouteLen + x.Onion.Len() }
func (x *Route) Wrap(inner Onion) { x.Onion = inner }

type Route struct {
	HiddenService *pub.Key
	HiddenCloaked cloak.PubKey
	Sender        *prv.Key
	SenderPub     *pub.Key
	nonce.IV
	// ------- the rest is encrypted to the HiddenService/Sender keys.
	ID nonce.ID
	// Ciphers is a set of 3 symmetric ciphers that are to be used in their
	// given order over the reply message from the service.
	Ciphers Ciphers
	// Nonces are the nonces to use with the cipher when creating the
	// encryption for the reply message,
	// they are common with the crypts in the header.
	Nonces
	RoutingHeaderBytes
	Onion
}

func (o Skins) Route(id nonce.ID, k *pub.Key, ks *signer.KeySet,
	ep *ExitPoint) Skins {
	
	oo := &Route{
		HiddenService: k,
		Sender:        ks.Next(),
		IV:            nonce.New(),
		ID:            id,
		Ciphers:       GenCiphers(ep.Keys, ep.ReturnPubs),
		Nonces:        ep.Nonces,
		Onion:         &End{},
	}
	oo.SenderPub = pub.Derive(oo.Sender)
	oo.HiddenCloaked = cloak.GetCloak(k)
	return append(o, oo)
}

func (x *Route) Encode(s *Splice) (e error) {
	log.T.S("encoding", reflect.TypeOf(x),
		x.ID, x.HiddenService, x.Sender, x.IV, x.Ciphers, x.Nonces,
		x.RoutingHeaderBytes,
	)
	s.Magic(RouteMagic).
		Cloak(x.HiddenService).
		Pubkey(pub.Derive(x.Sender)).
		IV(x.IV)
	start := s.GetCursor()
	s.ID(x.ID).Ciphers(x.Ciphers).Nonces(x.Nonces)
	if e = x.Onion.Encode(s); check(e) {
		return
	}
	var blk cipher.Block
	// Encrypt the message!
	if blk = ciph.GetBlock(x.Sender, x.HiddenService, "route"); check(e) {
		return
	}
	ciph.Encipher(blk, x.IV, s.GetFrom(start))
	return
}

func (x *Route) Decode(s *Splice) (e error) {
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
func (x *Route) Decrypt(prk *prv.Key, s *Splice) {
	// log.D.S(s.GetRange(-1, s.GetCursor()), s.GetRange(s.GetCursor(), -1))
	ciph.Encipher(ciph.GetBlock(prk, x.SenderPub, "route decrypt"), x.IV,
		s.GetCursorToEnd())
	// And now we can see the reply field for the return trip.
	s.ReadID(&x.ID).ReadCiphers(&x.Ciphers).ReadNonces(&x.Nonces).
		ReadRoutingHeader(&x.RoutingHeaderBytes)
}

func (x *Route) Handle(s *Splice, p Onion, ng *Engine) (e error) {
	
	log.D.Ln(ng.GetLocalNodeAddressString(), "handling route")
	hc := ng.FindCloakedHiddenService(x.HiddenCloaked)
	if hc == nil {
		log.T.Ln("no matching hidden service key found from cloaked key")
		return
	}
	x.HiddenService, e = pub.FromBytes((*hc)[:])
	if check(e) {
		return
	}
	log.D.Ln("route key", *hc)
	hcl := *hc
	if hh, ok := ng.HiddenRouting.HiddenServices[hcl]; ok {
		log.D.F("we are the hidden service %s - decrypting...",
			hh.CurrentIntros[0].Key.ToBase32Abbreviated())
		// We have the keys to unwrap this one.
		// log.D.Ln(s)
		x.Decrypt(hh.Prv, s)
		log.D.Ln(s)
		// Add another two hops for security against unmasking.
		// preHops := []byte{0, 1}
		// path := make(Sessions, 2)
		// ng.SelectHops(preHops, path, "route prehops")
		n := GenNonces(5)
		rvKeys := ng.KeySet.Next3()
		hops := []byte{3, 4, 5, 0, 1}
		sessions := make(Sessions, len(hops))
		ng.SelectHops(hops, sessions, "route reply header")
		rt := &Routing{
			Sessions: [3]*SessionData{sessions[0], sessions[1], sessions[2]},
			Keys:     Privs{rvKeys[0], rvKeys[1], rvKeys[2]},
			Nonces:   Nonces{n[0], n[1], n[2]},
		}
		rh := Skins{}.RoutingHeader(rt)
		rHdr := Encode(rh.Assemble())
		rHdr.SetCursor(0)
		ep := ExitPoint{
			Routing: rt,
			ReturnPubs: Pubs{
				pub.Derive(sessions[0].PayloadPrv),
				pub.Derive(sessions[1].PayloadPrv),
				pub.Derive(sessions[2].PayloadPrv),
			},
		}
		mr := Skins{}.
			ForwardCrypt(sessions[3], ng.KeySet.Next(), n[3]).
			ForwardCrypt(sessions[4], ng.KeySet.Next(), n[4]).
			Ready(x.ID, x.HiddenService,
				x.RoutingHeaderBytes,
				rHdr.GetRoutingHeaderFromCursor(),
				x.Ciphers,
				GenCiphers(ep.Keys, ep.ReturnPubs),
				x.Nonces,
				ep.Nonces)
		// log.D.S("makeready", mr)
		assembled := mr.Assemble()
		// log.D.S("assembled", assembled)
		reply := Encode(assembled)
		ng.HandleMessage(reply, x)
	}
	return
}

func MakeRoute(id nonce.ID, k *pub.Key, ks *signer.KeySet,
	alice, bob *SessionData, c Circuit) Skins {
	
	headers := GetHeaders(alice, bob, c, ks)
	// log.T.S("headers", headers)
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
	log.D.Ln(ng.GetLocalNodeAddressString(), "sending route",
		k.ToBase32Abbreviated())
	hops := StandardCircuit()
	s := make(Sessions, len(hops))
	s[2] = ss
	// log.D.S("sessions before", s)
	se := ng.SelectHops(hops, s, "sendroute")
	var c Circuit
	copy(c[:], se)
	// log.D.S("sessions after", c)
	o := MakeRoute(nonce.NewID(), k, ng.KeySet, se[5], c[2], c)
	// log.D.S("doing accounting", o)
	res := ng.PostAcctOnion(o)
	log.D.Ln("sending out route request onion")
	ng.SendWithOneHook(c[0].Node.AddrPort, res, hook, ng.PendingResponses)
}

func (x *Route) Account(res *SendData, sm *SessionManager, s *SessionData, last bool) (skip bool, sd *SessionData) {
	
	copy(res.ID[:], x.ID[:])
	res.Billable = append(res.Billable, s.ID)
	return
}
