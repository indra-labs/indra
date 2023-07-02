// Package session provides an onion message type that delivers the two session private keys to be associated with a session, for which the hash of the secrets was used as the payment preimage for starting a session.
//
// Topping up sessions does not require following up with this message as the handler finds the session and adjusts the balance according to the payment.
package session

import (
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/crypto/sha256"
	"github.com/indra-labs/indra/pkg/engine/coding"
	"github.com/indra-labs/indra/pkg/engine/magic"
	"github.com/indra-labs/indra/pkg/engine/sess"
	"github.com/indra-labs/indra/pkg/engine/sessions"
	"github.com/indra-labs/indra/pkg/onions/end"
	"github.com/indra-labs/indra/pkg/onions/ont"
	"github.com/indra-labs/indra/pkg/onions/reg"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/splice"
	"reflect"
)

var (
	log   = log2.GetLogger()
	fails = log.E.Chk
)

const (
	SessionMagic = "sess"
	SessionLen   = magic.Len + nonce.IDLen + crypto.PrvKeyLen*2
)

type Session struct {
	ID              nonce.ID // only used by a client
	Hop             byte     // only used by a client
	Header, Payload *crypto.Keys
	ont.Onion
}

func NewSession(hop byte) (x ont.Onion) {
	var e error
	var hdr, pld *crypto.Keys
	if hdr, pld, e = crypto.Generate2Keys(); fails(e) {
		return
	}
	return &Session{
		ID:      nonce.NewID(),
		Hop:     hop,
		Header:  hdr,
		Payload: pld,
		Onion:   end.NewEnd(),
	}
}

func (x *Session) Account(res *sess.Data, sm *sess.Manager, s *sessions.Data,
	last bool) (skip bool, sd *sessions.Data) {
	return
}

func (x *Session) Decode(s *splice.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), SessionLen-magic.Len,
		SessionMagic); fails(e) {
		return
	}
	var h, p crypto.Prv
	hdr, pld := &h, &p
	s.
		ReadID(&x.ID).
		ReadPrvkey(&hdr).
		ReadPrvkey(&pld)
	x.Header, x.Payload = crypto.MakeKeys(hdr), crypto.MakeKeys(pld)
	return
}

func (x *Session) Encode(s *splice.Splice) (e error) {
	log.T.S("encoding", reflect.TypeOf(x),
		x.ID, x.Hop, x.Header, x.Payload,
	)
	return x.Onion.Encode(s.Magic(SessionMagic).
		ID(x.ID).
		Prvkey(x.Header.Prv).
		Prvkey(x.Payload.Prv),
	)
}

func (x *Session) GetOnion() interface{} { return x }

func (x *Session) Handle(s *splice.Splice, p ont.Onion, ng ont.Ngin) (e error) {
	log.T.F("incoming session %s", x.PreimageHash())
	pi := ng.Mgr().FindPendingPreimage(x.PreimageHash())
	if pi != nil {
		// We need to delete this first in case somehow two such messages arrive
		// at the same time, and we end up with duplicate
		ng.Mgr().DeletePendingPayment(pi.Preimage)
		log.D.F("adding session %s to %s", pi.ID,
			ng.Mgr().GetLocalNodeAddressString())
		ng.Mgr().AddSession(sessions.NewSessionData(pi.ID,
			ng.Mgr().GetLocalNode(), pi.Amount, x.Header, x.Payload, x.Hop))
		ng.HandleMessage(splice.BudgeUp(s), nil)
	} else {
		log.E.Ln("dropping session message without payment")
	}
	return
}

func (x *Session) Len() int      { return SessionLen + x.Onion.Len() }
func (x *Session) Magic() string { return SessionMagic }

func (x *Session) PreimageHash() sha256.Hash {
	h, p := x.Header.Prv.ToBytes(), x.Payload.Prv.ToBytes()
	return sha256.Single(append(h[:], p[:]...))
}

func (x *Session) Wrap(inner ont.Onion) { x.Onion = inner }
func init()                             { reg.Register(SessionMagic, sessionGen) }
func sessionGen() coding.Codec          { return &Session{} }
