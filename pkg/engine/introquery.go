package engine

import (
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

func (o Skins) IntroQuery(hsk *pub.Key, point *ExitPoint) Skins {
	
	return append(o, &IntroQuery{
		Key:     hsk,
		Ciphers: GenCiphers(point.Keys, point.ReturnPubs),
		Nonces:  point.Nonces,
	})
}

func MakeIntroQuery(hsk *pub.Key, client *SessionData, c Circuit,
	ks *signer.KeySet) Skins {
	
	headers := GetHeaders(client, c, ks)
	return Skins{}.
		RoutingHeader(headers.Forward).
		IntroQuery(hsk, headers.ExitPoint()).
		RoutingHeader(headers.Return)
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
