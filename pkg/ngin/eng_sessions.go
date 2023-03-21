package ngin

import (
	"fmt"
	
	"git-indra.lan/indra-labs/lnd/lnd/lnwire"
	"github.com/gookit/color"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/util/cryptorand"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

// PostAcctOnion takes a slice of Skins and calculates their costs and
// the list of sessions inside them and attaches accounting operations to
// apply when the associated confirmation(s) or response hooks are executed.
func (sm *SessionManager) PostAcctOnion(o Skins) (res SendData) {
	assembled := o.Assemble()
	// log.T.S(assembled)
	res.B = Encode(assembled).GetRange(-1, -1)
	// do client accounting
	skip := false
	for i := range o {
		if skip {
			skip = false
			continue
		}
		switch on := o[i].(type) {
		case *Crypt:
			s := sm.FindSessionByHeaderPub(on.ToHeaderPub)
			if s == nil {
				continue
			}
			res.Sessions = append(res.Sessions, s)
			// The last hop needs no accounting as it's us!
			if i == len(o)-1 {
				// The session used for the last hop is stored, however.
				res.Ret = s.ID
				res.Billable = append(res.Billable, s.ID)
				break
			}
			switch on2 := o[i+1].(type) {
			case *Exit:
				for j := range s.Node.Services {
					if s.Node.Services[j].Port != on2.Port {
						continue
					}
					res.Port = on2.Port
					res.PostAcct = append(res.PostAcct,
						func() {
							sm.DecSession(s.ID,
								s.Node.Services[j].RelayRate*len(res.B)/2, true, "exit")
						})
					break
				}
				res.Billable = append(res.Billable, s.ID)
				res.Last = on2.ID
				skip = true
			case *Forward:
				res.Billable = append(res.Billable, s.ID)
				res.PostAcct = append(res.PostAcct,
					func() {
						sm.DecSession(s.ID, s.Node.RelayRate*len(res.B),
							true, "forward")
					})
			case *GetBalance:
				res.Last = s.ID
				res.Billable = append(res.Billable, s.ID)
				skip = true
			case *HiddenService:
				res.Last = on2.Intro.ID
				res.Billable = append(res.Billable, s.ID)
				skip = true
			case *Intro:
				log.D.Ln("intro in crypt")
				res.Last = on2.ID
			case *IntroQuery:
				res.Last = on2.ID
				res.Billable = append(res.Billable, s.ID)
				skip = true
			case *Reverse:
				res.Billable = append(res.Billable, s.ID)
			case *Route:
				copy(res.Last[:], on2.ID[:])
				res.Billable = append(res.Billable, s.ID)
			}
		case *Confirmation:
			res.Last = on.ID
		case *Balance:
			res.Last = on.ID
		case *Intro:
			log.D.Ln("intro not in crypt")
			res.Last = on.ID
		case *IntroQuery:
			log.D.Ln("introquery not in crypt")
			res.Last = on.ID
		}
	}
	return
}

// BuyNewSessions performs the initial purchase of 5 sessions as well as adding
// different hop numbers to relays with existing  Note that all 5 of
// the sessions will be paid the amount specified, not divided up.
func (ng *Engine) BuyNewSessions(amount lnwire.MilliSatoshi,
	fn func()) (e error) {
	
	var nodes [5]*Node
	nodes = ng.SessionManager.SelectUnusedCircuit()
	for i := range nodes {
		if nodes[i] == nil {
			e = fmt.Errorf("failed to find nodes %d", i)
			return
		}
	}
	// Get a random return hop session (index 5).
	var returnSession *SessionData
	returnHops := ng.SessionManager.GetSessionsAtHop(5)
	if len(returnHops) > 1 {
		cryptorand.Shuffle(len(returnHops), func(i, j int) {
			returnHops[i], returnHops[j] = returnHops[j], returnHops[i]
		})
	}
	// There must be at least one, and if there was more than one the first
	// index of returnHops will be a randomly selected one.
	returnSession = returnHops[0]
	conf := nonce.NewID()
	var s [5]*Session
	for i := range s {
		s[i] = NewSessionKeys(byte(i))
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
	o := MakeSession(conf, s, returnSession, nodes[:], ng.KeySet)
	res := ng.PostAcctOnion(o)
	ng.SendWithOneHook(nodes[0].AddrPort, res, func(id nonce.ID, k *pub.Bytes,
		b slice.Bytes) (e error) {
		ng.SessionManager.Lock()
		defer ng.SessionManager.Unlock()
		var ss [5]*SessionData
		for i := range nodes {
			log.D.F("confirming and storing session at hop %d %s for %s with"+
				" %v initial balance",
				i, s[i].ID,
				color.Yellow.Sprint(nodes[i].AddrPort.String()),
				amount)
			ss[i] = NewSessionData(s[i].ID, nodes[i], amount,
				s[i].Header, s[i].Payload, byte(i))
			ng.SessionManager.Add(ss[i])
			ng.Sessions = append(ng.Sessions, ss[i])
			ng.SessionManager.PendingPayments.Delete(s[i].PreimageHash())
		}
		fn()
		return
	}, ng.PendingResponses)
	return
}
