package relay

import (
	"fmt"
	
	"git-indra.lan/indra-labs/lnd/lnd/lnwire"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/onion"
	"git-indra.lan/indra-labs/indra/pkg/onion/layers/session"
	"git-indra.lan/indra-labs/indra/pkg/traffic"
	"git-indra.lan/indra-labs/indra/pkg/util/cryptorand"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

// BuyNewSessions performs the initial purchase of 5 sessions as well as adding
// different hop numbers to relays with existing sessions. Note that all 5 of
// the sessions will be paid the amount specified, not divided up.
func (eng *Engine) BuyNewSessions(amount lnwire.MilliSatoshi,
	hook func()) (e error) {
	
	var nodes [5]*traffic.Node
	nodes = eng.SessionManager.SelectUnusedCircuit()
	for i := range nodes {
		if nodes[i] == nil {
			e = fmt.Errorf("failed to find nodes %d", i)
			return
		}
	}
	// Get a random return hop session (index 5).
	var returnSession *traffic.Session
	returnHops := eng.SessionManager.GetSessionsAtHop(5)
	if len(returnHops) > 1 {
		cryptorand.Shuffle(len(returnHops), func(i, j int) {
			returnHops[i], returnHops[j] = returnHops[j], returnHops[i]
		})
	}
	// There must be at least one, and if there was more than one the first
	// index of returnHops will be a randomly selected one.
	returnSession = returnHops[0]
	conf := nonce.NewID()
	var s [5]*session.Layer
	for i := range s {
		s[i] = session.New(byte(i))
	}
	var confirmChans [5]chan bool
	var pendingConfirms int
	for i := range nodes {
		confirmChans[i] = nodes[i].
			PaymentChan.Send(amount, s[i])
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
	o := onion.SendKeys(conf, s, returnSession, nodes[:], eng.KeySet)
	res := eng.PostAcctOnion(o)
	eng.SendWithOneHook(nodes[0].AddrPort, res, 0,
		func(id nonce.ID, b slice.Bytes) {
			eng.SessionManager.Lock()
			defer eng.SessionManager.Unlock()
			var sessions [5]*traffic.Session
			for i := range nodes {
				log.D.F("confirming and storing session at hop %d %s for %s with"+
					" %v initial"+
					" balance", i, s[i].ID, nodes[i].AddrPort.String(), amount)
				sessions[i] = traffic.NewSession(s[i].ID, nodes[i], amount,
					s[i].Header, s[i].Payload, byte(i))
				eng.SessionManager.Add(sessions[i])
				eng.Sessions = append(eng.Sessions, sessions[i])
				eng.SessionManager.PendingPayments.Delete(s[i].PreimageHash())
			}
			hook()
		})
	return
}
