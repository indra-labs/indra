package relay

import (
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/signer"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/messages/balance"
	"git-indra.lan/indra-labs/indra/pkg/messages/crypt"
	"git-indra.lan/indra-labs/indra/pkg/messages/getbalance"
	"git-indra.lan/indra-labs/indra/pkg/types"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

func (eng *Engine) balance(on *balance.Layer,
	b slice.Bytes, c *slice.Cursor, prev types.Onion) {
	
	local := eng.GetLocalNodeAddress()
	pending := eng.PendingResponses.Find(on.ID)
	if pending != nil {
		for i := range pending.Billable {
			s := eng.FindSession(pending.Billable[i])
			if s != nil {
				switch {
				case i < 2:
					in := s.RelayRate *
						pending.SentSize
					eng.DecSession(s.ID, in, true, "reverse")
				case i == 2:
					in := s.RelayRate * pending.SentSize / 2
					out := s.RelayRate * len(b) / 2
					eng.DecSession(s.ID, in+out, true, "getbalance")
				case i > 2:
					out := s.RelayRate * len(b)
					eng.DecSession(s.ID, out, true, "reverse")
				}
			}
		}
		var se *Session
		eng.IterateSessions(func(s *Session) bool {
			if s.ID == on.ID {
				log.D.F("%s received balance %s for session %s %s was %s",
					local,
					on.MilliSatoshi, on.ID, on.ConfID, s.Remaining)
				se = s
				return true
			}
			return false
		})
		eng.PendingResponses.Delete(pending.ID, nil)
		if se != nil {
			log.D.F("got %v, expected %v", se.Remaining, on.MilliSatoshi)
			se.Remaining = on.MilliSatoshi
		}
	}
}

// GetBalance sends out a request in a similar way to SendExit except the node
// being queried can be any of the 5.
func GetBalance(id, confID nonce.ID, client *Session,
	s Circuit, ks *signer.KeySet) Skins {
	
	var prvs [3]*prv.Key
	for i := range prvs {
		prvs[i] = ks.Next()
	}
	n := GenNonces(6)
	var retNonces [3]nonce.IV
	copy(retNonces[:], n[3:])
	var pubs [3]*pub.Key
	pubs[0] = s[3].PayloadPub
	pubs[1] = s[4].PayloadPub
	pubs[2] = client.PayloadPub
	return Skins{}.
		ReverseCrypt(s[0], ks.Next(), n[0], 3).
		ReverseCrypt(s[1], ks.Next(), n[1], 2).
		ReverseCrypt(s[2], ks.Next(), n[2], 1).
		GetBalance(id, confID, prvs, pubs, retNonces).
		ReverseCrypt(s[3], prvs[0], n[3], 0).
		ReverseCrypt(s[4], prvs[1], n[4], 0).
		ReverseCrypt(client, prvs[2], n[5], 0)
}

func (eng *Engine) SendGetBalance(target *Session, hook Callback) {
	hops := []byte{0, 1, 2, 3, 4, 5}
	s := make(Sessions, len(hops))
	s[2] = target
	se := eng.SelectHops(hops, s)
	var c Circuit
	copy(c[:], se)
	confID := nonce.NewID()
	o := GetBalance(target.ID, confID, se[5], c, eng.KeySet)
	log.D.Ln("sending out getbalance onion")
	res := eng.PostAcctOnion(o)
	eng.SendWithOneHook(c[0].AddrPort, res, hook)
}

func (eng *Engine) getbalance(on *getbalance.Layer,
	b slice.Bytes, c *slice.Cursor, prev types.Onion) {
	
	log.T.S(on)
	var found bool
	var bal *balance.Layer
	eng.IterateSessions(func(s *Session) bool {
		if s.ID == on.ID {
			bal = &balance.Layer{
				ID:           on.ID,
				ConfID:       on.ConfID,
				MilliSatoshi: s.Remaining,
			}
			found = true
			return true
		}
		return false
	})
	if !found {
		log.E.Ln("session not found", on.ID)
		log.D.S(eng.Sessions)
		return
	}
	header := b[*c:c.Inc(crypt.ReverseHeaderLen)]
	rb := FormatReply(header,
		Encode(bal), on.Ciphers, on.Nonces)
	rb = append(rb, slice.NoisePad(714-len(rb))...)
	switch on1 := prev.(type) {
	case *crypt.Layer:
		sess := eng.FindSessionByHeader(on1.ToPriv)
		if sess != nil {
			in := sess.RelayRate * len(b) / 2
			out := sess.RelayRate * len(rb) / 2
			eng.DecSession(sess.ID, in+out, false, "getbalance")
		}
	}
	eng.IterateSessions(func(s *Session) bool {
		if s.ID == on.ID {
			bal = &balance.Layer{
				ID:           on.ID,
				ConfID:       on.ConfID,
				MilliSatoshi: s.Remaining,
			}
			found = true
			return true
		}
		return false
	})
	rb = FormatReply(header,
		Encode(bal), on.Ciphers, on.Nonces)
	rb = append(rb, slice.NoisePad(714-len(rb))...)
	eng.handleMessage(rb, on)
}
