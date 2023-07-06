// Package balance provides an onion layer message that comes in response to a getbalance query, informing the client of the balance of a session, identified by the getbalance nonce.ID.
package balance

import (
	"github.com/indra-labs/indra/pkg/codec"
	"github.com/indra-labs/indra/pkg/codec/ont"
	"github.com/indra-labs/indra/pkg/codec/reg"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/engine/magic"
	"github.com/indra-labs/indra/pkg/engine/sess"
	"github.com/indra-labs/indra/pkg/engine/sessions"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/slice"
	"github.com/indra-labs/indra/pkg/util/splice"
	"github.com/lightningnetwork/lnd/lnwire"
	"reflect"
)

var (
	log   = log2.GetLogger()
	fails = log.E.Chk
)

const (
	Magic = "bala"
	Len   = magic.Len + nonce.IDLen + slice.Uint64Len
)

// Balance is the balance with an ID the client has a pending balance update
// waiting on.
type Balance struct {

	// ID of the request this response relates to.
	ID nonce.ID

	// Amount current in the balance according to the relay.
	lnwire.MilliSatoshi
}

// Account simply records the message ID, which will be recognised in the pending
// responses cache.
func (x *Balance) Account(res *sess.Data, sm *sess.Manager,
	s *sessions.Data, last bool) (skip bool, sd *sessions.Data) {
	res.ID = x.ID
	return
}

// Decode a splice.Splice's next bytes into a Balance.
func (x *Balance) Decode(s *splice.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), Len-magic.Len,
		Magic); fails(e) {
		return
	}
	s.ReadID(&x.ID).
		ReadMilliSatoshi(&x.MilliSatoshi)
	return
}

// Encode a Balance into a splice.Splice's next bytes.
func (x *Balance) Encode(s *splice.Splice) (e error) {
	log.T.S("encoding", reflect.TypeOf(x),
		x.ID,
		x.MilliSatoshi,
	)
	s.
		Magic(Magic).
		ID(x.ID).
		Uint64(uint64(x.MilliSatoshi))
	return
}

// GetOnion returns nothing because there isn't an onion inside a Balance.
func (x *Balance) GetOnion() interface{} { return nil }

// Handle provides relay and accounting processing logic for receiving a Balance
// message.
func (x *Balance) Handle(s *splice.Splice, p ont.Onion, ng ont.Ngin) (e error) {
	log.D.Ln("handling balance", x.ID)
	log.D.S("pending", ng.Pending())
	if pending := ng.Pending().Find(x.ID); pending != nil {
		log.D.S("found pending", pending.ID)
		for i := range pending.Billable {
			session := ng.Mgr().FindSessionByPubkey(pending.Billable[i])
			out := int(session.Node.RelayRate) * s.Len()
			if session != nil {
				in := int(session.Node.RelayRate) * pending.SentSize
				shb := session.Header.Bytes
				switch {
				case i < 2:
					ng.Mgr().DecSession(shb, in, true, "reverse")
				case i == 2:
					ng.Mgr().DecSession(shb, (in+out)/2, true, "getbalance")
				case i > 2:
					ng.Mgr().DecSession(shb, out, true, "reverse")
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

// Len returns the length of bytes required to encode the Balance.
func (x *Balance) Len() int { return Len }

// Magic bytes that identify this message.
func (x *Balance) Magic() string { return Magic }

// Wrap is a no-op because a Balance is terminal.
func (x *Balance) Wrap(inner ont.Onion) {}

// New creates a new Balance as an ont.Onion.
func New(id nonce.ID, amt lnwire.MilliSatoshi) ont.Onion {
	return &Balance{ID: id, MilliSatoshi: amt}
}

// Gen is a factory function to generate an Ad.
func Gen() codec.Codec { return &Balance{} }

func init() { reg.Register(Magic, Gen) }
