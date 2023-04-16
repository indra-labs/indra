package engine

import (
	"reflect"
	
	"git-indra.lan/indra-labs/lnd/lnd/lnwire"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/engine/magic"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

const (
	BalanceMagic = "ba"
	BalanceLen   = magic.Len + nonce.IDLen*2 + slice.Uint64Len
)

type Balance struct {
	ID     nonce.ID
	ConfID nonce.ID
	lnwire.MilliSatoshi
}

func balanceGen() Codec             { return &Balance{} }
func init()                         { Register(BalanceMagic, balanceGen) }
func (x *Balance) Magic() string    { return BalanceMagic }
func (x *Balance) Len() int         { return BalanceLen }
func (x *Balance) Wrap(inner Onion) {}
func (x *Balance) GetOnion() Onion  { return x }

func (o Skins) Balance(id, confID nonce.ID,
	amt lnwire.MilliSatoshi) Skins {
	
	return append(o, &Balance{
		ID:           id,
		ConfID:       confID,
		MilliSatoshi: amt,
	})
}

func (x *Balance) Encode(s *Splice) (e error) {
	log.T.S("encoding", reflect.TypeOf(x),
		x.ID, x.ConfID, x.MilliSatoshi,
	)
	s.
		Magic(BalanceMagic).
		ID(x.ID).
		ID(x.ConfID).
		Uint64(uint64(x.MilliSatoshi))
	return
}

func (x *Balance) Decode(s *Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), BalanceLen-magic.Len,
		BalanceMagic); fails(e) {
		return
	}
	s.
		ReadID(&x.ID).
		ReadID(&x.ConfID).
		ReadMilliSatoshi(&x.MilliSatoshi)
	return
}

func (x *Balance) Handle(s *Splice, p Onion,
	ng *Engine) (e error) {
	
	if pending := ng.PendingResponses.Find(x.ID); pending != nil {
		log.D.S("found pending", pending.ID)
		for i := range pending.Billable {
			session := ng.FindSession(pending.Billable[i])
			out := session.Node.RelayRate * s.Len()
			if session != nil {
				in := session.Node.RelayRate * pending.SentSize
				switch {
				case i < 2:
					ng.DecSession(session.ID, in, true, "reverse")
				case i == 2:
					ng.DecSession(session.ID, (in+out)/2, true, "getbalance")
				case i > 2:
					ng.DecSession(session.ID, out, true, "reverse")
				}
			}
		}
		var se *SessionData
		ng.IterateSessions(func(s *SessionData) bool {
			if s.ID == x.ID {
				log.D.F("received balance %s for session %s %s was %s",
					x.MilliSatoshi, x.ID, x.ConfID, s.Remaining)
				se = s
				return true
			}
			return false
		})
		if se != nil {
			log.D.F("got %v, expected %v", se.Remaining, x.MilliSatoshi)
			se.Remaining = x.MilliSatoshi
		}
		ng.PendingResponses.ProcessAndDelete(pending.ID, nil, s.GetAll())
	}
	return
}

func (x *Balance) Account(res *SendData, sm *SessionManager, s *SessionData, last bool) (skip bool, sd *SessionData) {
	
	res.ID = x.ID
	return
}
