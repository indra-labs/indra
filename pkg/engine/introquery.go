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
	IntroQueryLen   = magic.Len + nonce.IDLen + pub.KeyLen +
		3*sha256.Len + nonce.IVLen*3
)

type IntroQuery struct {
	ID  nonce.ID
	Key *pub.Key
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

func (o Skins) IntroQuery(id nonce.ID, hsk *pub.Key, exit *ExitPoint) Skins {
	return append(o, &IntroQuery{
		ID:      id,
		Key:     hsk,
		Ciphers: GenCiphers(exit.Keys, exit.ReturnPubs),
		Nonces:  exit.Nonces,
	})
}

func (x *IntroQuery) Magic() string { return IntroQueryMagic }

func (x *IntroQuery) Encode(s *octet.Splice) (e error) {
	return x.Onion.Encode(s.
		Magic(IntroQueryMagic).
		ID(x.ID).
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
		ReadID(&x.ID).
		ReadPubkey(&x.Key).
		ReadHashTriple(&x.Ciphers).
		ReadIVTriple(&x.Nonces)
	return
}

func (x *IntroQuery) Len() int { return IntroQueryLen + x.Onion.Len() }

func (x *IntroQuery) Wrap(inner Onion) { x.Onion = inner }

func (x *IntroQuery) Handle(s *octet.Splice, p Onion,
	ng *Engine) (e error) {
	
	ng.Introductions.Lock()
	var ok bool
	var il *Intro
	if il, ok = ng.Introductions.KnownIntros[x.Key.ToBytes()]; !ok {
		// if the reply is zeroes the querant knows it needs to retry at a
		// different relay
		// il = &Intro{}
		ng.Introductions.Unlock()
		log.E.Ln("intro not known")
		return
	}
	ng.Introductions.Unlock()
	log.D.S(il.ID, il.Key, il.Sig)
	iq := Encode(il)
	rb := FormatReply(s.GetRange(s.GetCursor(), s.Advance(ReverseHeaderLen)),
		iq.GetRange(-1, -1), x.Ciphers, x.Nonces)
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

func MakeIntroQuery(id nonce.ID, hsk *pub.Key, target *SessionData, c Circuit,
	ks *signer.KeySet) Skins {
	
	headers := GetHeaders(target, c, ks)
	return Skins{}.
		RoutingHeader(headers.Forward).
		IntroQuery(id, hsk, headers.ExitPoint()).
		RoutingHeader(headers.Return)
}

func (ng *Engine) SendIntroQuery(id nonce.ID, hsk *pub.Key, target *SessionData,
	fn Callback) {
	
	hops := StandardCircuit()
	s := make(Sessions, len(hops))
	s[2] = target
	se := ng.SelectHops(hops, s)
	var c Circuit
	copy(c[:], se)
	o := MakeIntroQuery(id, hsk, s[5], c, ng.KeySet)
	// log.D.S(ng.GetLocalNodeAddress().String()+" sending out intro query onion",
	// 	o)
	res := ng.PostAcctOnion(o)
	// res.Last = id
	ng.SendWithOneHook(c[0].AddrPort, res, fn, ng.PendingResponses)
}
