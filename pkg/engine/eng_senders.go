package engine

import (
	"net/netip"
	"time"

	"github.com/gookit/color"

	"git-indra.lan/indra-labs/indra/pkg/crypto"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/engine/onions"
	"git-indra.lan/indra-labs/indra/pkg/engine/responses"
	"git-indra.lan/indra-labs/indra/pkg/engine/services"
	"git-indra.lan/indra-labs/indra/pkg/engine/sess"
	"git-indra.lan/indra-labs/indra/pkg/engine/sessions"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
	"git-indra.lan/indra-labs/indra/pkg/util/splice"
)

func (ng *Engine) SendExit(port uint16, msg slice.Bytes, id nonce.ID,
	alice, bob *sessions.Data, hook responses.Callback) {

	hops := sess.StandardCircuit()
	s := make(sessions.Sessions, len(hops))
	s[2] = bob
	s[5] = alice
	se := ng.Manager.SelectHops(hops, s, "exit")
	var c sessions.Circuit
	copy(c[:], se)
	o := onions.MakeExit(onions.ExitParams{port, msg, id, bob, alice, c, ng.KeySet})
	res := PostAcctOnion(ng.Manager, o)
	ng.Manager.SendWithOneHook(c[0].Node.AddrPort, res, hook, ng.Responses)
}

func (ng *Engine) SendGetBalance(alice, bob *sessions.Data, hook responses.Callback) {
	hops := sess.StandardCircuit()
	s := make(sessions.Sessions, len(hops))
	s[2] = bob
	s[5] = alice
	se := ng.Manager.SelectHops(hops, s, "sendgetbalance")
	var c sessions.Circuit
	copy(c[:], se)
	o := onions.MakeGetBalance(onions.GetBalanceParams{alice.ID, alice, bob, c,
		ng.KeySet})
	log.D.Ln("sending out getbalance onion")
	res := PostAcctOnion(ng.Manager, o)
	ng.Manager.SendWithOneHook(c[0].Node.AddrPort, res, hook, ng.Responses)
}

func (ng *Engine) SendHiddenService(id nonce.ID, key *crypto.Prv,
	relayRate uint32, port uint16, expiry time.Time,
	alice, bob *sessions.Data, svc *services.Service,
	hook responses.Callback) (in *onions.Intro) {

	hops := sess.StandardCircuit()
	s := make(sessions.Sessions, len(hops))
	s[2] = alice
	se := ng.Manager.SelectHops(hops, s, "sendhiddenservice")
	var c sessions.Circuit
	copy(c[:], se[:len(c)])
	in = onions.NewIntro(id, key, alice.Node.AddrPort, relayRate, port, expiry)
	o := onions.MakeHiddenService(in, alice, bob, c, ng.KeySet)
	log.D.F("%s sending out hidden service onion %s",
		ng.Manager.GetLocalNodeAddressString(),
		color.Yellow.Sprint(alice.Node.AddrPort.String()))
	res := PostAcctOnion(ng.Manager, o)
	ng.GetHidden().AddHiddenService(svc, key, in,
		ng.Manager.GetLocalNodeAddressString())
	ng.Manager.SendWithOneHook(c[0].Node.AddrPort, res, hook, ng.Responses)
	return
}

func (ng *Engine) SendIntroQuery(id nonce.ID, hsk *crypto.Pub,
	alice, bob *sessions.Data, hook func(in *onions.Intro)) {

	fn := func(id nonce.ID, ifc interface{}, b slice.Bytes) (e error) {
		s := splice.Load(b, slice.NewCursor())
		on := onions.Recognise(s)
		if e = on.Decode(s); fails(e) {
			return
		}
		var oni *onions.Intro
		var ok bool
		if oni, ok = on.(*onions.Intro); !ok {
			return
		}
		hook(oni)
		return
	}
	log.D.Ln("sending introquery")
	hops := sess.StandardCircuit()
	s := make(sessions.Sessions, len(hops))
	s[2] = bob
	s[5] = alice
	se := ng.Manager.SelectHops(hops, s, "sendintroquery")
	var c sessions.Circuit
	copy(c[:], se)
	o := onions.MakeIntroQuery(id, hsk, bob, alice, c, ng.KeySet)
	res := PostAcctOnion(ng.Manager, o)
	log.D.Ln(res.ID)
	ng.Manager.SendWithOneHook(c[0].Node.AddrPort, res, fn, ng.Responses)
}

func (ng *Engine) SendMessage(mp *onions.Message, hook responses.Callback) (id nonce.ID) {
	// Add another two hops for security against unmasking.
	preHops := []byte{0, 1}
	oo := ng.Manager.SelectHops(preHops, mp.Forwards[:], "sendmessage")
	mp.Forwards = [2]*sessions.Data{oo[0], oo[1]}
	o := onions.Skins{}.Message(mp, ng.KeySet)
	res := PostAcctOnion(ng.Manager, o)
	log.D.Ln("sending out message onion")
	ng.Manager.SendWithOneHook(mp.Forwards[0].Node.AddrPort, res, hook,
		ng.Responses)
	return res.ID
}

func (ng *Engine) SendPing(c sessions.Circuit, hook responses.Callback) {
	hops := sess.StandardCircuit()
	s := make(sessions.Sessions, len(hops))
	copy(s, c[:])
	se := ng.Manager.SelectHops(hops, s, "sendping")
	copy(c[:], se)
	id := nonce.NewID()
	o := onions.Ping(id, se[len(se)-1], c, ng.KeySet)
	res := PostAcctOnion(ng.Manager, o)
	ng.Manager.SendWithOneHook(c[0].Node.AddrPort, res, hook, ng.Responses)
}

func (ng *Engine) SendRoute(k *crypto.Pub, ap *netip.AddrPort,
	hook responses.Callback) {

	ng.Manager.FindNodeByAddrPort(ap)
	var ss *sessions.Data
	ng.Manager.IterateSessions(func(s *sessions.Data) bool {
		if s.Node.AddrPort.String() == ap.String() {
			ss = s
			return true
		}
		return false
	})
	if ss == nil {
		log.E.Ln(ng.Manager.GetLocalNodeAddressString(),
			"could not find session for address", ap.String())
		return
	}
	log.D.Ln(ng.Manager.GetLocalNodeAddressString(), "sending route",
		k.ToBase32Abbreviated())
	hops := sess.StandardCircuit()
	s := make(sessions.Sessions, len(hops))
	s[2] = ss
	se := ng.Manager.SelectHops(hops, s, "sendroute")
	var c sessions.Circuit
	copy(c[:], se)
	o := onions.MakeRoute(nonce.NewID(), k, ng.KeySet, se[5], c[2], c)
	res := PostAcctOnion(ng.Manager, o)
	log.D.Ln("sending out route request onion")
	ng.Manager.SendWithOneHook(c[0].Node.AddrPort, res, hook, ng.Responses)
}
