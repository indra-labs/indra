package engine

import (
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/signer"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/engine/magic"
	"git-indra.lan/indra-labs/indra/pkg/util/octet"
)

const (
	IntroQueryMagic = "iq"
	IntroQueryLen   = magic.Len + pub.KeyLen +
		3*sha256.Len + nonce.IVLen*3
)

type IntroQuery struct {
	*pub.Key
	// Ciphers is a set of 3 symmetric ciphers that are to be used in their
	// given order over the reply message from the service.
	Ciphers [3]sha256.Hash
	// Nonces are the nonces to use with the cipher when creating the
	// encryption for the reply message.
	Nonces [3]nonce.IV
	Onion
}

func introQueryPrototype() Onion { return &IntroQuery{} }

func init() { Register(IntroQueryMagic, introQueryPrototype) }

func (o Skins) IntroQuery(hsk *pub.Key, prvs [3]*prv.Key, pubs [3]*pub.Key,
	nonces [3]nonce.IV) Skins {
	
	return append(o, &IntroQuery{
		Key:     hsk,
		Ciphers: GenCiphers(prvs, pubs),
		Nonces:  nonces,
	})
}

func MakeIntroQuery(hsk *pub.Key, client *SessionData, c Circuit,
	ks *signer.KeySet) Skins {
	
	forwardKeys := ks.Next3()
	returnKeys := ks.Next3()
	n := GenNonces(6)
	var returnNonces, forwardNonces [3]nonce.IV
	copy(returnNonces[:], n[3:])
	copy(forwardNonces[:], n[:3])
	var forwardSessions, returnSessions [3]*SessionData
	copy(forwardSessions[:], c[:3])
	copy(returnSessions[:], c[3:5])
	returnSessions[2] = client
	var returnPubs [3]*pub.Key
	returnPubs[0] = c[3].PayloadPub
	returnPubs[1] = c[4].PayloadPub
	returnPubs[2] = client.PayloadPub
	return Skins{}.
		RoutingHeader(forwardSessions, forwardKeys, forwardNonces).
		IntroQuery(hsk, returnKeys, returnPubs, returnNonces).
		RoutingHeader(returnSessions, returnKeys, returnNonces)
}

func (x *IntroQuery) Magic() string { return IntroQueryMagic }

func (x *IntroQuery) Encode(s *octet.Splice) (e error) {
	return x.Onion.Encode(s.
		Magic(IntroQueryMagic).
		Pubkey(x.Key).
		HashTriple(x.Ciphers).
		IVTriple(x.Nonces),
	)
}

func (x *IntroQuery) Decode(s *octet.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), IntroQueryLen-magic.Len,
		IntroQueryMagic); check(e) {
		return
	}
	s.
		ReadPubkey(&x.Key).
		ReadHashTriple(&x.Ciphers).
		ReadIVTriple(&x.Nonces)
	return
}

func (x *IntroQuery) Len() int { return IntroQueryLen }

func (x *IntroQuery) Wrap(inner Onion) { x.Onion = inner }

func (x *IntroQuery) Handle(s *octet.Splice, p Onion,
	ng *Engine) (e error) {
	
	ng.Introductions.Lock()
	var ok bool
	var il *Intro
	if il, ok = ng.Introductions.KnownIntros[x.Key.ToBytes()]; !ok {
		// if the reply is zeroes the querant knows it needs to retry at a
		// different relay
		il = &Intro{}
	}
	ng.Introductions.Unlock()
	rb := FormatReply(s.GetRange(s.GetCursor(), s.Advance(ReverseHeaderLen)),
		Encode(il).GetRange(-1, -1), x.Ciphers, x.Nonces)
	switch on1 := p.(type) {
	case *Crypt:
		sess := ng.FindSessionByHeader(on1.ToPriv)
		if sess != nil {
			in := sess.RelayRate * s.Len() / 2
			out := sess.RelayRate * rb.Len() / 2
			ng.DecSession(sess.ID, in+out, false, "introquery")
		}
	}
	ng.HandleMessage(rb, x)
	return
}
