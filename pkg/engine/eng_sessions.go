package engine

import (
	"fmt"
	"github.com/gookit/color"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/engine/node"
	"github.com/indra-labs/indra/pkg/engine/onions"
	"github.com/indra-labs/indra/pkg/engine/sessions"
	"github.com/indra-labs/indra/pkg/util/cryptorand"
	"github.com/indra-labs/indra/pkg/util/slice"
	"github.com/lightningnetwork/lnd/lnwire"
)

// BuyNewSessions performs the initial purchase of 5 sessions as well as adding
// different hop numbers to relays with existing  Note that all 5 of
// the sessions will be paid the amount specified, not divided up.
func (ng *Engine) BuyNewSessions(amount lnwire.MilliSatoshi,
	fn func()) (e error) {
	var nodes [5]*node.Node
	nodes = ng.Manager.SelectUnusedCircuit()
	for i := range nodes {
		if nodes[i] == nil {
			e = fmt.Errorf("failed to find nodes %d", i)
			return
		}
	}
	// Get a random return hop session (index 5).
	var returnSession *sessions.Data
	returnHops := ng.Manager.GetSessionsAtHop(5)
	if len(returnHops) > 1 {
		cryptorand.Shuffle(len(returnHops), func(i, j int) {
			returnHops[i], returnHops[j] = returnHops[j], returnHops[i]
		})
	}
	// There must be at least one, and if there was more than one the first
	// index of returnHops will be a randomly selected one.
	returnSession = returnHops[0]
	conf := nonce.NewID()
	var s [5]*onions.Session
	for i := range s {
		s[i] = onions.NewSessionKeys(byte(i))
	}
	var confirmChans [5]chan bool
	var pendingConfirms int
	for i := range nodes {
		confirmChans[i] = nodes[i].
			Chan.Send(amount, s[i].ID, s[i].PreimageHash())
		pendingConfirms++
	}
	var success bool
	for pendingConfirms > 0 {
		// The confirmation channels will signal upon success or failure
		// according to the LN payment send protocol once either the HTLCs
		// confirm on the way back or the path fails.
		select {
		case success = <-confirmChans[0]:
			if success {
				pendingConfirms--
			}
		case success = <-confirmChans[1]:
			if success {
				pendingConfirms--
			}
		case success = <-confirmChans[2]:
			if success {
				pendingConfirms--
			}
		case success = <-confirmChans[3]:
			if success {
				pendingConfirms--
			}
		case success = <-confirmChans[4]:
			if success {
				pendingConfirms--
			}
		}
	}
	// todo: handle payment failures!
	o := onions.MakeSession(conf, s, returnSession, nodes[:], ng.KeySet)
	res := PostAcctOnion(ng.Manager, o)
	ng.Manager.SendWithOneHook(nodes[0].AddrPort, res, func(id nonce.ID,
		ifc interface{},
		b slice.Bytes) (e error) {
		ng.Manager.Lock()
		defer ng.Manager.Unlock()
		var ss [5]*sessions.Data
		for i := range nodes {
			log.D.F("confirming and storing session at hop %d %s for %s with"+
				" %v initial balance",
				i, s[i].ID,
				color.Yellow.Sprint(nodes[i].AddrPort.String()),
				amount)
			ss[i] = sessions.NewSessionData(s[i].ID, nodes[i], amount,
				s[i].Header, s[i].Payload, byte(i))
			ng.Manager.Add(ss[i])
			ng.Manager.Sessions = append(ng.Manager.Sessions, ss[i])
			ng.Manager.PendingPayments.Delete(s[i].PreimageHash())
		}
		fn()
		return
	}, ng.Responses)
	return
}
