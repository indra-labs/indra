package engine

import (
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/signer"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/engine/magic"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
	"git-indra.lan/indra-labs/indra/pkg/util/zip"
)

const (
	IntroQueryMagic = "iq"
	IntroQueryLen   = magic.Len + nonce.IDLen + pub.KeyLen +
		3*sha256.Len + nonce.IVLen*3
)

type IntroQuery struct {
	zip.Reply
	Key *pub.Key
	Onion
}

func introQueryPrototype() Onion { return &IntroQuery{} }

func init() { Register(IntroQueryMagic, introQueryPrototype) }

func (o Skins) IntroQuery(id nonce.ID, hsk *pub.Key, exit *ExitPoint) Skins {
	return append(o, &IntroQuery{
		Reply: zip.Reply{
			ID:      id,
			Ciphers: GenCiphers(exit.Keys, exit.ReturnPubs),
			Nonces:  exit.Nonces,
		},
		Key:   hsk,
		Onion: nop,
	})
}

func (x *IntroQuery) Magic() string { return IntroQueryMagic }

func (x *IntroQuery) Encode(s *zip.Splice) (e error) {
	return x.Onion.Encode(s.
		Magic(IntroQueryMagic).
		Reply(&x.Reply).
		Pubkey(x.Key),
	)
}

func (x *IntroQuery) Decode(s *zip.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), IntroQueryLen-magic.Len,
		IntroQueryMagic); check(e) {
		return
	}
	s.ReadReply(&x.Reply).ReadPubkey(&x.Key)
	return
}

func (x *IntroQuery) Len() int { return IntroQueryLen + x.Onion.Len() }

func (x *IntroQuery) Wrap(inner Onion) { x.Onion = inner }

func (x *IntroQuery) Handle(s *zip.Splice, p Onion,
	ng *Engine) (e error) {
	
	ng.HiddenRouting.Lock()
	var ok bool
	var il *Intro
	if il, ok = ng.HiddenRouting.KnownIntros[x.Key.ToBytes()]; !ok {
		// if the reply is zeroes the querant knows it needs to retry at a
		// different relay
		il = &Intro{}
		ng.HiddenRouting.Unlock()
		log.E.Ln("intro not known")
		return
	}
	ng.HiddenRouting.Unlock()
	// log.D.S(il.ID, il.Key, il.Expiry, il.Sig)
	iqr := Encode(il)
	rb := FormatReply(s.GetRange(s.GetCursor(), s.Advance(ReverseHeaderLen)),
		iqr.GetRange(-1, -1), x.Ciphers, x.Nonces)
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

func (ng *Engine) SendIntroQuery(id nonce.ID, hsk *pub.Key,
	target *SessionData, hook func(in *Intro)) {
	
	fn := func(id nonce.ID, k *pub.Bytes, b slice.Bytes) (e error) {
		s := zip.Load(b, slice.NewCursor())
		on := Recognise(s)
		if e = on.Decode(s); check(e) {
			return
		}
		var oni *Intro
		var ok bool
		if oni, ok = on.(*Intro); !ok {
			return
		}
		hook(oni)
		return
	}
	
	hops := StandardCircuit()
	s := make(Sessions, len(hops))
	s[2] = target
	se := ng.SelectHops(hops, s)
	var c Circuit
	copy(c[:], se)
	o := MakeIntroQuery(id, hsk, s[5], c, ng.KeySet)
	res := ng.PostAcctOnion(o)
	ng.SendWithOneHook(c[0].AddrPort, res, fn, ng.PendingResponses)
}
