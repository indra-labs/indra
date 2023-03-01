package relay

import (
	"encoding/hex"
	"fmt"
	
	"git-indra.lan/indra-labs/lnd/lnd/lnwire"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/signer"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/relay/messages/session"
	"git-indra.lan/indra-labs/indra/pkg/relay/types"
	"git-indra.lan/indra-labs/indra/pkg/util/cryptorand"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

// A Session keeps track of a connection session. It specifically maintains the
// account of available bandwidth allocation before it needs to be recharged
// with new credit, and the current state of the encryption.
type Session struct {
	ID nonce.ID
	*Node
	Remaining                 lnwire.MilliSatoshi
	HeaderPrv, PayloadPrv     *prv.Key
	HeaderPub, PayloadPub     *pub.Key
	HeaderBytes, PayloadBytes pub.Bytes
	Preimage                  sha256.Hash
	Hop                       byte
}

// NewSession creates a new Session, generating cached public key bytes and
// preimage.
func NewSession(
	id nonce.ID,
	node *Node,
	rem lnwire.MilliSatoshi,
	hdrPrv *prv.Key,
	pldPrv *prv.Key,
	hop byte,
) (s *Session) {
	
	var e error
	if hdrPrv == nil || pldPrv == nil {
		if hdrPrv, e = prv.GenerateKey(); check(e) {
		}
		if pldPrv, e = prv.GenerateKey(); check(e) {
		}
	}
	hdrPub := pub.Derive(hdrPrv)
	pldPub := pub.Derive(pldPrv)
	h, p := hdrPrv.ToBytes(), pldPrv.ToBytes()
	s = &Session{
		ID:           id,
		Node:         node,
		Remaining:    rem,
		HeaderPub:    hdrPub,
		HeaderBytes:  hdrPub.ToBytes(),
		PayloadPub:   pldPub,
		PayloadBytes: pldPub.ToBytes(),
		HeaderPrv:    hdrPrv,
		PayloadPrv:   pldPrv,
		Preimage:     sha256.Single(append(h[:], p[:]...)),
		Hop:          hop,
	}
	return
}

// IncSats adds to the Remaining counter, used when new data allowance has been
// purchased.
func (s *Session) IncSats(sats lnwire.MilliSatoshi, sender bool, typ string) {
	who := "relay"
	if sender {
		who = "client"
	}
	log.D.F("%s session %d %x current %v incrementing by %v", who, typ, s.ID,
		s.Remaining, sats)
	s.Remaining += sats
}

// DecSats reduces the amount Remaining, if the requested amount would put
// the total below zero it returns false, signalling that new data allowance
// needs to be purchased before any further messages can be sent.
func (s *Session) DecSats(sats lnwire.MilliSatoshi, sender bool,
	typ string) bool {
	
	if s.Remaining < sats {
		return false
	}
	who := "relay"
	if sender {
		who = "client"
	}
	log.D.F("%s session %s %s current %v decrementing by %v", who, typ, s.ID,
		s.Remaining, sats)
	s.Remaining -= sats
	return true
}

// A Circuit is the generic fixed length path used for most messages.
type Circuit [5]*Session

func (c Circuit) String() (o string) {
	o += "[ "
	for i := range c {
		if c[i] == nil {
			o += "                 "
		} else {
			o += hex.EncodeToString(c[i].ID[:]) + " "
		}
	}
	o += "]"
	return
}

// Sessions are arbitrary length lists of sessions.
type Sessions []*Session

func (eng *Engine) session(on *session.Layer, b slice.Bytes,
	c *slice.Cursor, prev types.Onion) {
	
	log.D.Ln(prev == nil)
	log.T.F("incoming session %x", on.PreimageHash())
	pi := eng.FindPendingPreimage(on.PreimageHash())
	if pi != nil {
		// We need to delete this first in case somehow two such messages arrive
		// at the same time, and we end up with duplicate sessions.
		eng.DeletePendingPayment(pi.Preimage)
		log.D.F("Adding session %s to %s", pi.ID, eng.GetLocalNodeAddress())
		eng.AddSession(NewSession(pi.ID,
			eng.GetLocalNode(), pi.Amount, on.Header, on.Payload, on.Hop))
		eng.handleMessage(BudgeUp(b, *c), on)
	} else {
		log.E.Ln("dropping session message without payment")
	}
}

// BuyNewSessions performs the initial purchase of 5 sessions as well as adding
// different hop numbers to relays with existing sessions. Note that all 5 of
// the sessions will be paid the amount specified, not divided up.
func (eng *Engine) BuyNewSessions(amount lnwire.MilliSatoshi,
	fn func()) (e error) {
	
	var nodes [5]*Node
	nodes = eng.SessionManager.SelectUnusedCircuit()
	for i := range nodes {
		if nodes[i] == nil {
			e = fmt.Errorf("failed to find nodes %d", i)
			return
		}
	}
	// Get a random return hop session (index 5).
	var returnSession *Session
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
	o := SendSessions(conf, s, returnSession, nodes[:], eng.KeySet)
	res := eng.PostAcctOnion(o)
	eng.SendWithOneHook(nodes[0].AddrPort, res, func(id nonce.ID, b slice.Bytes) {
		eng.SessionManager.Lock()
		defer eng.SessionManager.Unlock()
		var sessions [5]*Session
		for i := range nodes {
			log.D.F("confirming and storing session at hop %d %s for %s with"+
				" %v initial"+
				" balance", i, s[i].ID, nodes[i].AddrPort.String(), amount)
			sessions[i] = NewSession(s[i].ID, nodes[i], amount,
				s[i].Header, s[i].Payload, byte(i))
			eng.SessionManager.Add(sessions[i])
			eng.Sessions = append(eng.Sessions, sessions[i])
			eng.SessionManager.PendingPayments.Delete(s[i].PreimageHash())
		}
		fn()
	})
	return
}

// SendSessions provides a pair of private keys that will be used to generate the
// Purchase header bytes and to generate the ciphers provided in the Purchase
// message to encrypt the Session that is returned.
//
// The OnionSkin key, its cloaked public key counterpart used in the ToHeaderPub
// field of the Purchase message preformed header bytes, but the Ciphers
// provided in the Purchase message, for encrypting the Session to be returned,
// uses the Payload key, along with the public key found in the encrypted crypt
// of the header for the Reverse relay.
//
// This message's last crypt is a Confirmation, which allows the client to know
// that the keys were successfully delivered.
//
// This is the only onion that uses the node identity keys. The payment preimage
// hash must be available or the relay should not forward the remainder of the
// packet.
//
// If hdr/pld cipher keys are nil there must be a HeaderPub available on the
// session for the hop. This allows this function to send keys to any number of
// hops, but the very first SendSessions must have all in order to create the first
// set of sessions. This is by way of indicating to not use the IdentityPub but
// the HeaderPub instead. Not allowing free relay at all prevents spam attacks.
func SendSessions(id nonce.ID, s [5]*session.Layer,
	client *Session, hop []*Node, ks *signer.KeySet) Skins {
	
	n := GenNonces(6)
	sk := Skins{}
	for i := range s {
		if i == 0 {
			sk = sk.Crypt(hop[i].IdentityPub, nil, ks.Next(),
				n[i], 0).Session(s[i])
		} else {
			sk = sk.ForwardSession(hop[i], ks.Next(), n[i], s[i])
		}
	}
	return sk.
		ForwardCrypt(client, ks.Next(), n[5]).
		Confirmation(id, 0)
}
