package engine

import (
	"net/netip"
	"time"
	
	"github.com/gookit/color"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/splice"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

func (ng *Engine) SendExit(port uint16, msg slice.Bytes, id nonce.ID,
	alice, bob *SessionData, hook Callback) {
	
	hops := StandardCircuit()
	s := make(Sessions, len(hops))
	s[2] = bob
	s[5] = alice
	se := ng.SelectHops(hops, s, "exit")
	var c Circuit
	copy(c[:], se)
	o := MakeExit(ExitParams{port, msg, id, bob, alice, c, ng.KeySet})
	res := ng.PostAcctOnion(o)
	ng.SendWithOneHook(c[0].Node.AddrPort, res, hook, ng.PendingResponses)
}

func (ng *Engine) SendGetBalance(alice, bob *SessionData, hook Callback) {
	hops := StandardCircuit()
	s := make(Sessions, len(hops))
	s[2] = bob
	s[5] = alice
	se := ng.SelectHops(hops, s, "sendgetbalance")
	var c Circuit
	copy(c[:], se)
	confID := nonce.NewID()
	o := MakeGetBalance(GetBalanceParams{alice.ID, confID, alice, bob, c,
		ng.KeySet})
	log.D.Ln("sending out getbalance onion")
	res := ng.PostAcctOnion(o)
	ng.SendWithOneHook(c[0].Node.AddrPort, res, hook, ng.PendingResponses)
}

func (ng *Engine) SendHiddenService(id nonce.ID, key *crypto.Prv,
	expiry time.Time, alice, bob *SessionData,
	svc *Service, hook Callback) (in *Intro) {
	
	hops := StandardCircuit()
	s := make(Sessions, len(hops))
	s[2] = alice
	se := ng.SelectHops(hops, s, "sendhiddenservice")
	var c Circuit
	copy(c[:], se[:len(c)])
	in = NewIntro(id, key, alice.Node.AddrPort, expiry)
	o := MakeHiddenService(in, alice, bob, c, ng.KeySet)
	log.D.F("%s sending out hidden service onion %s",
		ng.GetLocalNodeAddressString(),
		color.Yellow.Sprint(alice.Node.AddrPort.String()))
	res := ng.PostAcctOnion(o)
	ng.HiddenRouting.AddHiddenService(svc, key, in,
		ng.GetLocalNodeAddressString())
	ng.SendWithOneHook(c[0].Node.AddrPort, res, hook, ng.PendingResponses)
	return
}

func (ng *Engine) SendIntroQuery(id nonce.ID, hsk *crypto.Pub,
	alice, bob *SessionData, hook func(in *Intro)) {
	
	fn := func(id nonce.ID, ifc interface{}, b slice.Bytes) (e error) {
		s := splice.Load(b, slice.NewCursor())
		on := Recognise(s)
		if e = on.Decode(s); fails(e) {
			return
		}
		var oni *Intro
		var ok bool
		if oni, ok = on.(*Intro); !ok {
			return
		}
		hook(oni)
		return
	}
	log.D.Ln("sending introquery")
	hops := StandardCircuit()
	s := make(Sessions, len(hops))
	s[2] = bob
	s[5] = alice
	se := ng.SelectHops(hops, s, "sendintroquery")
	var c Circuit
	copy(c[:], se)
	o := MakeIntroQuery(id, hsk, bob, alice, c, ng.KeySet)
	res := ng.PostAcctOnion(o)
	log.D.Ln(res.ID)
	ng.SendWithOneHook(c[0].Node.AddrPort, res, fn, ng.PendingResponses)
}

func (ng *Engine) SendMessage(mp *Message, hook Callback) (id nonce.ID) {
	// Add another two hops for security against unmasking.
	preHops := []byte{0, 1}
	oo := ng.SelectHops(preHops, mp.Forwards[:], "sendmessage")
	mp.Forwards = [2]*SessionData{oo[0], oo[1]}
	o := Skins{}.Message(mp, ng.KeySet)
	res := ng.PostAcctOnion(o)
	log.D.Ln("sending out message onion")
	ng.SendWithOneHook(mp.Forwards[0].Node.AddrPort, res, hook,
		ng.PendingResponses)
	return res.ID
}

func (ng *Engine) SendPing(c Circuit, hook Callback) {
	hops := StandardCircuit()
	s := make(Sessions, len(hops))
	copy(s, c[:])
	se := ng.SelectHops(hops, s, "sendping")
	copy(c[:], se)
	confID := nonce.NewID()
	o := Ping(confID, se[len(se)-1], c, ng.KeySet)
	res := ng.PostAcctOnion(o)
	ng.SendWithOneHook(c[0].Node.AddrPort, res, hook, ng.PendingResponses)
}

func (ng *Engine) SendRoute(k *crypto.Pub, ap *netip.AddrPort,
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
	se := ng.SelectHops(hops, s, "sendroute")
	var c Circuit
	copy(c[:], se)
	o := MakeRoute(nonce.NewID(), k, ng.KeySet, se[5], c[2], c)
	res := ng.PostAcctOnion(o)
	log.D.Ln("sending out route request onion")
	ng.SendWithOneHook(c[0].Node.AddrPort, res, hook, ng.PendingResponses)
}
