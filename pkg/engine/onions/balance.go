package onions

import (
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/engine/coding"
	"github.com/indra-labs/indra/pkg/engine/magic"
	"github.com/indra-labs/indra/pkg/engine/onions/reg"
	"github.com/indra-labs/indra/pkg/engine/sess"
	"github.com/indra-labs/indra/pkg/engine/sessions"
	"github.com/indra-labs/indra/pkg/util/slice"
	"github.com/indra-labs/indra/pkg/util/splice"
	"github.com/lightningnetwork/lnd/lnwire"
	"reflect"
)

const (
	BalanceMagic = "bala"
	BalanceLen   = magic.Len + nonce.IDLen + slice.Uint64Len
)

type Balance struct {
	ID nonce.ID
	lnwire.MilliSatoshi
}

func (x *Balance) Account(res *sess.Data, sm *sess.Manager,
	s *sessions.Data, last bool) (skip bool, sd *sessions.Data) {
	res.ID = x.ID
	return
}

func (x *Balance) Decode(s *splice.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), BalanceLen-magic.Len,
		BalanceMagic); fails(e) {
		return
	}
	s.ReadID(&x.ID).
		ReadMilliSatoshi(&x.MilliSatoshi)
	return
}

func (x *Balance) Encode(s *splice.Splice) (e error) {
	log.T.S("encoding", reflect.TypeOf(x),
		x.ID,
		x.MilliSatoshi,
	)
	s.
		Magic(BalanceMagic).
		ID(x.ID).
		Uint64(uint64(x.MilliSatoshi))
	return
}

func (x *Balance) GetOnion() interface{} { return x }

func (x *Balance) Handle(s *splice.Splice, p Onion, ng Ngin) (e error) {
	if pending := ng.Pending().Find(x.ID); pending != nil {
		log.D.S("found pending", pending.ID)
		for i := range pending.Billable {
			session := ng.Mgr().FindSession(pending.Billable[i])
			out := session.Node.RelayRate * s.Len()
			if session != nil {
				in := session.Node.RelayRate * pending.SentSize
				switch {
				case i < 2:
					ng.Mgr().DecSession(session.Header.Bytes, in, true, "reverse")
				case i == 2:
					ng.Mgr().DecSession(session.Header.Bytes, (in+out)/2, true, "getbalance")
				case i > 2:
					ng.Mgr().DecSession(session.Header.Bytes, out, true, "reverse")
				}
			}
		}
		var se *sessions.Data
		ng.Mgr().IterateSessions(func(s *sessions.Data) bool {
			if s.ID == x.ID {
				log.D.F("received balance %s for session %s was %s",
					x.MilliSatoshi, x.ID, s.Remaining)
				se = s
				return true
			}
			return false
		})
		if se != nil {
			log.D.F("got %v, expected %v", se.Remaining, x.MilliSatoshi)
			se.Remaining = x.MilliSatoshi
		}
		ng.Pending().ProcessAndDelete(pending.ID, nil, s.GetAll())
	}
	return
}

func (x *Balance) Len() int         { return BalanceLen }
func (x *Balance) Magic() string    { return BalanceMagic }
func (x *Balance) Wrap(inner Onion) {}
func balanceGen() coding.Codec      { return &Balance{} }
func init()                         { reg.Register(BalanceMagic, balanceGen) }
