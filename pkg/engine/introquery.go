package engine

import (
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/signer"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/engine/magic"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

const (
	IntroQueryMagic = "iq"
	IntroQueryLen   = magic.Len + nonce.IDLen + pub.KeyLen +
		3*sha256.Len + nonce.IVLen*3
)

type IntroQuery struct {
	Reply
	Key *pub.Key
	Onion
}

func introQueryPrototype() Onion { return &IntroQuery{} }

func init() { Register(IntroQueryMagic, introQueryPrototype) }

func (o Skins) IntroQuery(id nonce.ID, hsk *pub.Key, exit *ExitPoint) Skins {
	return append(o, &IntroQuery{
		Reply: Reply{
			ID:      id,
			Ciphers: GenCiphers(exit.Keys, exit.ReturnPubs),
			Nonces:  exit.Nonces,
		},
		Key:   hsk,
		Onion: nop,
	})
}

func (x *IntroQuery) Magic() string { return IntroQueryMagic }

func (x *IntroQuery) Encode(s *Splice) (e error) {
	// log.T.S("encoding", reflect.TypeOf(x),
	// 	x.FwReply, x.Key,
	// )
	return x.Onion.Encode(s.
		Magic(IntroQueryMagic).
		Reply(&x.Reply).
		Pubkey(x.Key),
	)
}

func (x *IntroQuery) Decode(s *Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), IntroQueryLen-magic.Len,
		IntroQueryMagic); check(e) {
		return
	}
	s.ReadReply(&x.Reply).ReadPubkey(&x.Key)
	return
}

func (x *IntroQuery) Len() int { return IntroQueryLen + x.Onion.Len() }

func (x *IntroQuery) Wrap(inner Onion) { x.Onion = inner }

func (x *IntroQuery) Handle(s *Splice, p Onion,
	ng *Engine) (e error) {
	
	ng.HiddenRouting.Lock()
	log.D.Ln(ng.GetLocalNodeAddressString(), "handling introquery", x.ID,
		x.Key.ToBase32Abbreviated())
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
	rb := FormatReply(s.GetRange(s.GetCursor(), s.Advance(RoutingHeaderLen,
		"routing header")), x.Ciphers, x.Nonces, iqr.GetAll())
	switch on1 := p.(type) {
	case *Crypt:
		sess := ng.FindSessionByHeader(on1.ToPriv)
		if sess != nil {
			in := sess.Node.RelayRate * s.Len() / 2
			out := sess.Node.RelayRate * rb.Len() / 2
			ng.DecSession(sess.ID, in+out, false, "introquery")
		}
	}
	ng.HandleMessage(rb, x)
	return
}

func MakeIntroQuery(id nonce.ID, hsk *pub.Key, alice, bob *SessionData,
	c Circuit, ks *signer.KeySet) Skins {
	
	headers := GetHeaders(alice, bob, c, ks)
	return Skins{}.
		RoutingHeader(headers.Forward).
		IntroQuery(id, hsk, headers.ExitPoint()).
		RoutingHeader(headers.Return)
}

func (ng *Engine) SendIntroQuery(id nonce.ID, hsk *pub.Key,
	alice, bob *SessionData, hook func(in *Intro)) {
	
	fn := func(id nonce.ID, ifc interface{}, b slice.Bytes) (e error) {
		// log.D.S("sendintroquery callback", id, k, b.ToBytes())
		s := Load(b, slice.NewCursor())
		on := Recognise(s, ng.GetLocalNodeAddress())
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
	log.D.Ln("sending introquery")
	hops := StandardCircuit()
	s := make(Sessions, len(hops))
	s[2] = bob
	s[5] = alice
	se := ng.SelectHops(hops, s, "sendintroquery")
	var c Circuit
	copy(c[:], se)
	o := MakeIntroQuery(id, hsk, bob, alice, c, ng.KeySet)
	res := ng.PostAcctOnion(o)
	log.D.Ln(res.ID)
	ng.SendWithOneHook(c[0].Node.AddrPort, res, fn, ng.PendingResponses)
}
