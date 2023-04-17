package engine

import (
	"reflect"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/engine/coding"
	"git-indra.lan/indra-labs/indra/pkg/engine/magic"
	"git-indra.lan/indra-labs/indra/pkg/engine/sessions"
	"git-indra.lan/indra-labs/indra/pkg/splice"
)

const (
	SessionMagic = "ss"
	SessionLen   = magic.Len + nonce.IDLen + crypto.PrvKeyLen*2
)

type Session struct {
	ID              nonce.ID // only used by a client
	Hop             byte     // only used by a client
	Header, Payload *crypto.Prv
	Onion
}

func sessionGen() coding.Codec           { return &Session{} }
func init()                              { Register(SessionMagic, sessionGen) }
func (x *Session) Magic() string         { return SessionMagic }
func (x *Session) Len() int              { return SessionLen + x.Onion.Len() }
func (x *Session) Wrap(inner Onion)      { x.Onion = inner }
func (x *Session) GetOnion() interface{} { return x }

func NewSessionKeys(hop byte) (x *Session) {
	var e error
	var hdrPrv, pldPrv *crypto.Prv
	if hdrPrv, e = crypto.GeneratePrvKey(); fails(e) {
		return
	}
	if pldPrv, e = crypto.GeneratePrvKey(); fails(e) {
		return
	}
	return &Session{
		ID:      nonce.NewID(),
		Hop:     hop,
		Header:  hdrPrv,
		Payload: pldPrv,
	}
}

func (x *Session) Encode(s *splice.Splice) (e error) {
	log.T.S("encoding", reflect.TypeOf(x),
		x.ID, x.Hop, x.Header, x.Payload,
	)
	return x.Onion.Encode(s.Magic(SessionMagic).
		ID(x.ID).
		Prvkey(x.Header).
		Prvkey(x.Payload),
	)
}

func (x *Session) Decode(s *splice.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), SessionLen-magic.Len,
		SessionMagic); fails(e) {
		return
	}
	s.
		ReadID(&x.ID).
		ReadPrvkey(&x.Header).
		ReadPrvkey(&x.Payload)
	return
}

func (x *Session) Handle(s *splice.Splice, p Onion, ni interface{}) (e error) {
	
	ng := ni.(*Engine)
	log.T.F("incoming session %s", x.PreimageHash())
	pi := ng.FindPendingPreimage(x.PreimageHash())
	if pi != nil {
		// We need to delete this first in case somehow two such messages arrive
		// at the same time, and we end up with duplicate 
		ng.DeletePendingPayment(pi.Preimage)
		log.D.F("adding session %s to %s", pi.ID, ng.GetLocalNodeAddressString())
		ng.AddSession(sessions.NewSessionData(pi.ID,
			ng.GetLocalNode(), pi.Amount, x.Header, x.Payload, x.Hop))
		ng.HandleMessage(splice.BudgeUp(s), nil)
	} else {
		log.E.Ln("dropping session message without payment")
	}
	return
}

func (x *Session) PreimageHash() sha256.Hash {
	h, p := x.Header.ToBytes(), x.Payload.ToBytes()
	return sha256.Single(append(h[:], p[:]...))
}

func (x *Session) Account(res *Data, sm *SessionManager, s *sessions.Data,
	last bool) (skip bool, sd *sessions.Data) {
	return
}
