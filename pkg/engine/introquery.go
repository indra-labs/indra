package engine

import (
	"reflect"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/engine/magic"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

const (
	IntroQueryMagic = "iq"
	IntroQueryLen   = magic.Len + nonce.IDLen + crypto.PubKeyLen +
		3*sha256.Len + nonce.IVLen*3
)

type IntroQuery struct {
	ID nonce.ID
	// Ciphers is a set of 3 symmetric ciphers that are to be used in their
	// given order over the reply message from the service.
	Ciphers
	// Nonces are the nonces to use with the cipher when creating the
	// encryption for the reply message,
	// they are common with the crypts in the header.
	Nonces
	// Port identifies the type of service as well as being the port used by
	// the service to be relayed to. Notice there is no IP address, this is
	// because Indranet only forwards to exits of decentralised services
	// also running on the same machine. This service could be a proxy, of
	// course, if configured this way. This could be done by tunneling from
	// a local Socks5 proxy into Indranet and the exit node also having
	// this.
	Key *crypto.Pub
	Onion
}

func introQueryGen() Codec             { return &IntroQuery{} }
func init()                            { Register(IntroQueryMagic, introQueryGen) }
func (x *IntroQuery) Magic() string    { return IntroQueryMagic }
func (x *IntroQuery) Len() int         { return IntroQueryLen + x.Onion.Len() }
func (x *IntroQuery) Wrap(inner Onion) { x.Onion = inner }
func (x *IntroQuery) GetOnion() Onion  { return x }

func (o Skins) IntroQuery(id nonce.ID, hsk *crypto.Pub, exit *ExitPoint) Skins {
	return append(o, &IntroQuery{
		ID:      id,
		Ciphers: GenCiphers(exit.Keys, exit.ReturnPubs),
		Nonces:  exit.Nonces,
		Key:     hsk,
		Onion:   nop,
	})
}

func (x *IntroQuery) Encode(s *Splice) (e error) {
	log.T.S("encoding", reflect.TypeOf(x),
		x.ID, x.Key, x.Ciphers, x.Nonces,
	)
	return x.Onion.Encode(s.
		Magic(IntroQueryMagic).
		ID(x.ID).Ciphers(x.Ciphers).Nonces(x.Nonces).
		Pubkey(x.Key),
	)
}

func (x *IntroQuery) Decode(s *Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), IntroQueryLen-magic.Len,
		IntroQueryMagic); fails(e) {
		return
	}
	s.ReadID(&x.ID).ReadCiphers(&x.Ciphers).ReadNonces(&x.Nonces).
		ReadPubkey(&x.Key)
	return
}

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
	iqr := Encode(il)
	rb := FormatReply(s.GetRoutingHeaderFromCursor(), x.Ciphers, x.Nonces,
		iqr.GetAll())
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

func MakeIntroQuery(id nonce.ID, hsk *crypto.Pub, alice, bob *SessionData,
	c Circuit, ks *crypto.KeySet) Skins {
	
	headers := GetHeaders(alice, bob, c, ks)
	return Skins{}.
		RoutingHeader(headers.Forward).
		IntroQuery(id, hsk, headers.ExitPoint()).
		RoutingHeader(headers.Return)
}

func (ng *Engine) SendIntroQuery(id nonce.ID, hsk *crypto.Pub,
	alice, bob *SessionData, hook func(in *Intro)) {
	
	fn := func(id nonce.ID, ifc interface{}, b slice.Bytes) (e error) {
		s := LoadSplice(b, slice.NewCursor())
		on := Recognise(s)
		if e = on.Decode(s); fails(e) {
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

func (x *IntroQuery) Account(res *SendData, sm *SessionManager, s *SessionData, last bool) (skip bool, sd *SessionData) {
	
	res.ID = x.ID
	res.Billable = append(res.Billable, s.ID)
	skip = true
	return
}
