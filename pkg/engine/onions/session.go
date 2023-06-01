package onions

import (
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/crypto/sha256"
	"github.com/indra-labs/indra/pkg/engine/coding"
	"github.com/indra-labs/indra/pkg/engine/magic"
	"github.com/indra-labs/indra/pkg/engine/onions/reg"
	"github.com/indra-labs/indra/pkg/engine/sess"
	"github.com/indra-labs/indra/pkg/engine/sessions"
	"github.com/indra-labs/indra/pkg/util/splice"
	"reflect"
)

const (
	SessionMagic = "sess"
	SessionLen   = magic.Len + nonce.IDLen + crypto.PrvKeyLen*2
)

type Session struct {
	ID              nonce.ID // only used by a client
	Hop             byte     // only used by a client
	Header, Payload *crypto.Keys
	Onion
}

func NewSessionKeys(hop byte) (x *Session) {
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

func (x *Session) Handle(s *splice.Splice, p Onion, ng Ngin) (e error) {
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

func (x *Session) Wrap(inner Onion) { x.Onion = inner }
func init()                         { reg.Register(SessionMagic, sessionGen) }
func sessionGen() coding.Codec      { return &Session{} }
