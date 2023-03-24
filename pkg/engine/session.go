package engine

import (
	"reflect"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/signer"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/engine/magic"
)

const (
	SessionMagic = "ss"
	SessionLen   = magic.Len + nonce.IDLen + prv.KeyLen*2
)

type Session struct {
	ID              nonce.ID // only used by a client
	Hop             byte     // only used by a client
	Header, Payload *prv.Key
	Onion
}

func sessionPrototype() Onion { return &Session{} }

func init() { Register(SessionMagic, sessionPrototype) }

func MakeSession(id nonce.ID, s [5]*Session,
	client *SessionData, hop []*Node, ks *signer.KeySet) Skins {
	
	n := GenNonces(6)
	sk := Skins{}
	for i := range s {
		if i == 0 {
			sk = sk.Crypt(hop[i].Identity.Pub, nil, ks.Next(),
				n[i], 0).Session(s[i])
		} else {
			sk = sk.ForwardSession(hop[i], ks.Next(), n[i], s[i])
		}
	}
	return sk.
		ForwardCrypt(client, ks.Next(), n[5]).
		Confirmation(id, 0)
}

func (o Skins) Session(sess *Session) Skins {
	// MakeSession can apply to from 1 to 5 nodes, if either key is nil then
	// this crypt just doesn't get added in the serialization process.
	if sess.Header == nil || sess.Payload == nil {
		return o
	}
	return append(o, &Session{
		Header:  sess.Header,
		Payload: sess.Payload,
		Onion:   &End{},
	})
}

func NewSessionKeys(hop byte) (x *Session) {
	var e error
	var hdrPrv, pldPrv *prv.Key
	if hdrPrv, e = prv.GenerateKey(); check(e) {
		return
	}
	if pldPrv, e = prv.GenerateKey(); check(e) {
		return
	}
	return &Session{
		ID:      nonce.NewID(),
		Hop:     hop,
		Header:  hdrPrv,
		Payload: pldPrv,
	}
}

func (x *Session) Magic() string { return SessionMagic }

func (x *Session) Encode(s *Splice) (e error) {
	log.T.S("encoding", reflect.TypeOf(x),
		x.ID, x.Hop, x.Header, x.Payload,
	)
	return x.Onion.Encode(s.Magic(SessionMagic).
		ID(x.ID).
		Prvkey(x.Header).
		Prvkey(x.Payload),
	)
}

func (x *Session) Decode(s *Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), SessionLen-magic.Len,
		SessionMagic); check(e) {
		return
	}
	s.
		ReadID(&x.ID).
		ReadPrvkey(&x.Header).
		ReadPrvkey(&x.Payload)
	return
}

func (x *Session) Len() int { return SessionLen + x.Onion.Len() }

func (x *Session) Wrap(inner Onion) { x.Onion = inner }

func (x *Session) Handle(s *Splice, p Onion, ng *Engine) (e error) {
	
	log.T.F("incoming session %s", x.PreimageHash())
	pi := ng.FindPendingPreimage(x.PreimageHash())
	if pi != nil {
		// We need to delete this first in case somehow two such messages arrive
		// at the same time, and we end up with duplicate 
		ng.DeletePendingPayment(pi.Preimage)
		log.D.F("adding session %s to %s", pi.ID, ng.GetLocalNodeAddressString())
		ng.AddSession(NewSessionData(pi.ID,
			ng.GetLocalNode(), pi.Amount, x.Header, x.Payload, x.Hop))
		ng.HandleMessage(BudgeUp(s), nil)
	} else {
		log.E.Ln("dropping session message without payment")
	}
	return
}

func (x *Session) PreimageHash() sha256.Hash {
	h, p := x.Header.ToBytes(), x.Payload.ToBytes()
	return sha256.Single(append(h[:], p[:]...))
}
