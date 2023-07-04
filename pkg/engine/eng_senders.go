package engine

import (
	"github.com/indra-labs/indra/pkg/onions/ad/intro"
	"github.com/indra-labs/indra/pkg/onions/exit"
	"github.com/indra-labs/indra/pkg/onions/getbalance"
	"github.com/indra-labs/indra/pkg/onions/message"
	"github.com/indra-labs/indra/pkg/onions/ont"
	"github.com/indra-labs/indra/pkg/onions/reg"
	"github.com/indra-labs/indra/pkg/util/multi"
	"net/netip"
	"time"

	"github.com/gookit/color"

	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/engine/responses"
	"github.com/indra-labs/indra/pkg/engine/services"
	"github.com/indra-labs/indra/pkg/engine/sess"
	"github.com/indra-labs/indra/pkg/engine/sessions"
	"github.com/indra-labs/indra/pkg/util/slice"
	"github.com/indra-labs/indra/pkg/util/splice"
)

// SendExit constructs and dispatches an exit message containing the desired message for the service at the exit.
func (ng *Engine) SendExit(port uint16, msg slice.Bytes, id nonce.ID,
	alice, bob *sessions.Data, hook responses.Callback) {

	hops := sess.StandardCircuit()
	s := make(sessions.Sessions, len(hops))
	s[2] = bob
	s[5] = alice
	se := ng.Mgr().SelectHops(hops, s, "exit")
	var c sessions.Circuit
	copy(c[:], se)
	o := MakeExit(exit.ExitParams{port, msg, id, bob, alice, c, ng.KeySet}, ng.Mgr().Protocols)
	res := PostAcctOnion(ng.Mgr(), o)
	ng.Mgr().SendWithOneHook(c[0].Node.Addresses, res, hook, ng.Responses)
}

// SendGetBalance sends out a balance request to a specific relay we have a session with.
func (ng *Engine) SendGetBalance(alice, bob *sessions.Data, hook responses.Callback) {
	hops := sess.StandardCircuit()
	s := make(sessions.Sessions, len(hops))
	s[2] = bob
	s[5] = alice
	se := ng.Mgr().SelectHops(hops, s, "sendgetbalance")
	var c sessions.Circuit
	copy(c[:], se)
	o := MakeGetBalance(getbalance.GetBalanceParams{alice.ID, alice, bob, c,
		ng.KeySet}, ng.Mgr().Protocols)
	log.D.S("sending out getbalance onion", o)
	res := PostAcctOnion(ng.Mgr(), o)
	ng.Mgr().SendWithOneHook(c[0].Node.Addresses, res, hook, ng.Responses)
}

// SendHiddenService dispatches a hiddenservice message, providing a relay the
// ability to refer clients to the hidden service and initiate connections.
func (ng *Engine) SendHiddenService(id nonce.ID, key *crypto.Prv, relayRate uint32,
	port uint16, expiry time.Time, alice, bob *sessions.Data, svc *services.Service,
	hook responses.Callback) (in *intro.Ad, e error) {

	hops := sess.StandardCircuit()
	s := make(sessions.Sessions, len(hops))
	s[2] = alice
	se := ng.Mgr().SelectHops(hops, s, "sendhiddenservice")
	var c sessions.Circuit
	copy(c[:], se[:len(c)])
	var addr netip.AddrPort
	if addr, e = multi.AddrToAddrPort(alice.Node.PickAddress(ng.Mgr().Protocols)); fails(e) {
		return
	}
	in = intro.New(id, key, &addr, relayRate, port, expiry)
	o := MakeHiddenService(in, alice, bob, c, ng.KeySet, ng.Mgr().Protocols)
	log.D.F("%s sending out hidden service onion %s",
		ng.Mgr().GetLocalNodeAddressString(),
		color.Yellow.Sprint(addr.String()))
	res := PostAcctOnion(ng.Mgr(), o)
	ng.GetHidden().AddHiddenService(svc, key, in,
		ng.Mgr().GetLocalNodeAddressString())
	ng.Mgr().SendWithOneHook(c[0].Node.Addresses, res, hook, ng.Responses)
	return
}

// SendIntroQuery delivers a query for a specified hidden service public key.
func (ng *Engine) SendIntroQuery(id nonce.ID, hsk *crypto.Pub,
	alice, bob *sessions.Data, hook func(in *intro.Ad)) {

	fn := func(id nonce.ID, ifc interface{}, b slice.Bytes) (e error) {
		s := splice.Load(b, slice.NewCursor())
		on := reg.Recognise(s)
		if e = on.Decode(s); fails(e) {
			return
		}
		var oni *intro.Ad
		var ok bool
		if oni, ok = on.(*intro.Ad); !ok {
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
	se := ng.Mgr().SelectHops(hops, s, "sendintroquery")
	var c sessions.Circuit
	copy(c[:], se)
	o := MakeIntroQuery(id, hsk, bob, alice, c, ng.KeySet, ng.Mgr().Protocols)
	res := PostAcctOnion(ng.Mgr(), o)
	log.D.Ln(res.ID)
	ng.Mgr().SendWithOneHook(c[0].Node.Addresses, res, fn, ng.Responses)
}

// SendMessage delivers a message to a hidden service for which we have an
// existing routing header from a previous or initial message cycle.
func (ng *Engine) SendMessage(mp *message.Message, hook responses.Callback) (id nonce.ID) {
	// Add another two hops for security against unmasking.
	preHops := []byte{0, 1}
	oo := ng.Mgr().SelectHops(preHops, mp.Forwards[:], "sendmessage")
	mp.Forwards = [2]*sessions.Data{oo[0], oo[1]}
	o := []ont.Onion{mp}
	res := PostAcctOnion(ng.Mgr(), o)
	log.D.Ln("sending out message onion")
	ng.Mgr().SendWithOneHook(mp.Forwards[0].Node.Addresses, res, hook,
		ng.Responses)
	return res.ID
}

// SendPing sends out a ping message, which provides some low-precision
// information about the online status and latency of a set of 5 peers we have
// sessions with.
func (ng *Engine) SendPing(c sessions.Circuit, hook responses.Callback) {
	hops := sess.StandardCircuit()
	s := make(sessions.Sessions, len(hops))
	copy(s, c[:])
	se := ng.Mgr().SelectHops(hops, s, "sendping")
	copy(c[:], se)
	id := nonce.NewID()
	o := Ping(id, se[len(se)-1], c, ng.KeySet, ng.Mgr().Protocols)
	res := PostAcctOnion(ng.Mgr(), o)
	ng.Mgr().SendWithOneHook(c[0].Node.Addresses, res, hook, ng.Responses)
}

// SendRoute delivers a message to establish a connection to a hidden service via
// an introducing relay found either through gossip or introquery.
func (ng *Engine) SendRoute(k *crypto.Pub, ap *netip.AddrPort,
	hook responses.Callback) {

	ng.Mgr().FindNodeByAddrPort(ap)
	var ss *sessions.Data
	ng.Mgr().IterateSessions(func(s *sessions.Data) bool {
		for _, v := range s.Node.Addresses {
			if v.String() == ap.String() {
				ss = s
				return true
			}
		}
		return false
	})
	if ss == nil {
		log.E.Ln(ng.Mgr().GetLocalNodeAddressString(),
			"could not find session for address", ap.String())
		return
	}
	log.D.Ln(ng.Mgr().GetLocalNodeAddressString(), "sending route",
		k.ToBased32Abbreviated())
	hops := sess.StandardCircuit()
	s := make(sessions.Sessions, len(hops))
	s[2] = ss
	se := ng.Mgr().SelectHops(hops, s, "sendroute")
	var c sessions.Circuit
	copy(c[:], se)
	o := MakeRoute(nonce.NewID(), k, ng.KeySet, se[5], c[2], c, ng.Mgr().Protocols)
	res := PostAcctOnion(ng.Mgr(), o)
	log.D.Ln("sending out route request onion")
	ng.Mgr().SendWithOneHook(c[0].Node.Addresses, res, hook, ng.Responses)
}
