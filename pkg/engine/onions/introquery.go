package onions

import (
	"git-indra.lan/indra-labs/indra/pkg/crypto"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/engine/coding"
	"git-indra.lan/indra-labs/indra/pkg/engine/magic"
	"git-indra.lan/indra-labs/indra/pkg/engine/sess"
	"git-indra.lan/indra-labs/indra/pkg/engine/sessions"
	"git-indra.lan/indra-labs/indra/pkg/util/splice"
	"reflect"
)

const (
	IntroQueryMagic = "intq"
	IntroQueryLen   = magic.Len + nonce.IDLen + crypto.PubKeyLen +
		3*sha256.Len + nonce.IVLen*3
)

type IntroQuery struct {
	ID nonce.ID
	// Ciphers is a set of 3 symmetric ciphers that are to be used in their
	// given order over the reply message from the service.
	crypto.Ciphers
	// Nonces are the nonces to use with the cipher when creating the
	// encryption for the reply message,
	// they are common with the crypts in the header.
	crypto.Nonces
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

func (x *IntroQuery) Account(res *sess.Data, sm *sess.Manager, s *sessions.Data, last bool) (skip bool, sd *sessions.Data) {
	res.ID = x.ID
	res.Billable = append(res.Billable, s.Header.Bytes)
	skip = true
	return
}

func (x *IntroQuery) Decode(s *splice.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), IntroQueryLen-magic.Len,
		IntroQueryMagic); fails(e) {
		return
	}
	s.ReadID(&x.ID).ReadCiphers(&x.Ciphers).ReadNonces(&x.Nonces).
		ReadPubkey(&x.Key)
	return
}

func (x *IntroQuery) Encode(s *splice.Splice) (e error) {
	log.T.S("encoding", reflect.TypeOf(x),
		x.ID, x.Key, x.Ciphers, x.Nonces,
	)
	return x.Onion.Encode(s.
		Magic(IntroQueryMagic).
		ID(x.ID).Ciphers(x.Ciphers).Nonces(x.Nonces).
		Pubkey(x.Key),
	)
}

func (x *IntroQuery) GetOnion() interface{} { return x }

func (x *IntroQuery) Handle(s *splice.Splice, p Onion, ng Ngin) (e error) {
	ng.GetHidden().Lock()
	log.D.Ln(ng.Mgr().GetLocalNodeAddressString(), "handling introquery", x.ID,
		x.Key.ToBased32Abbreviated())
	var ok bool
	var il *Intro
	if il, ok = ng.GetHidden().KnownIntros[x.Key.ToBytes()]; !ok {
		// if the reply is zeroes the querant knows it needs to retry at a
		// different relay
		il = &Intro{}
		ng.GetHidden().Unlock()
		log.E.Ln("intro not known")
		return
	}
	ng.GetHidden().Unlock()
	iqr := Encode(il)
	rb := FormatReply(GetRoutingHeaderFromCursor(s), x.Ciphers, x.Nonces,
		iqr.GetAll())
	switch on1 := p.(type) {
	case *Crypt:
		sess := ng.Mgr().FindSessionByHeader(on1.ToPriv)
		if sess != nil {
			in := sess.Node.RelayRate * s.Len() / 2
			out := sess.Node.RelayRate * rb.Len() / 2
			ng.Mgr().DecSession(sess.Header.Bytes, in+out, false, "introquery")
		}
	}
	ng.HandleMessage(rb, x)
	return
}

func (x *IntroQuery) Len() int         { return IntroQueryLen + x.Onion.Len() }
func (x *IntroQuery) Magic() string    { return IntroQueryMagic }
func (x *IntroQuery) Wrap(inner Onion) { x.Onion = inner }
func init()                            { Register(IntroQueryMagic, introQueryGen) }
func introQueryGen() coding.Codec { return &IntroQuery{} }
