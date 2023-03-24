package engine

import (
	"reflect"
	"time"
	
	"github.com/gookit/color"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/signer"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/engine/magic"
)

const (
	HiddenServiceMagic = "hs"
	HiddenServiceLen   = magic.Len + IntroLen +
		3*sha256.Len + nonce.IVLen*3 + RoutingHeaderLen
)

type HiddenService struct {
	Intro
	// Ciphers is a set of 3 symmetric ciphers that are to be used in their
	// given order over the reply message from the service.
	Ciphers
	// Nonces are the nonces to use with the cipher when creating the encryption
	// for the reply message, they are common with the crypts in the header.
	Nonces
	RoutingHeaderBytes
	Onion
}

func hiddenServicePrototype() Onion { return &HiddenService{} }

func init() { Register(HiddenServiceMagic, hiddenServicePrototype) }

func (o Skins) HiddenService(in *Intro, point *ExitPoint) Skins {
	return append(o, &HiddenService{
		Intro:   *in,
		Ciphers: GenCiphers(point.Keys, point.ReturnPubs),
		Nonces:  point.Nonces,
		Onion:   NewEnd(),
	})
}

func (x *HiddenService) Magic() string { return HiddenServiceMagic }

func (x *HiddenService) Encode(s *Splice) (e error) {
	log.T.S("encoding", reflect.TypeOf(x),
		x.ID, x.Key, x.AddrPort, x.Ciphers, x.Nonces, x.RoutingHeaderBytes,
	)
	SpliceIntro(s.Magic(HiddenServiceMagic), &x.Intro)
	return x.Onion.Encode(s.Ciphers(x.Ciphers).Nonces(x.Nonces))
}

func (x *HiddenService) Decode(s *Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), HiddenServiceLen-magic.Len,
		HiddenServiceMagic); check(e) {
		return
	}
	if e = x.Intro.Decode(s); check(e) {
		return
	}
	s.
		ReadCiphers(&x.Ciphers).
		ReadNonces(&x.Nonces).
		RoutingHeader(s.GetRoutingHeaderFromCursor())
	return
}

func (x *HiddenService) Len() int { return HiddenServiceLen + x.Onion.Len() }

func (x *HiddenService) Wrap(inner Onion) { x.Onion = inner }

func (x *HiddenService) Handle(s *Splice, p Onion, ng *Engine) (e error) {
	log.D.F("%s adding introduction for key %s",
		ng.GetLocalNodeAddressString(), x.Key.ToBase32Abbreviated())
	ng.HiddenRouting.AddIntro(x.Key, &Introduction{
		Intro:              &x.Intro,
		Ciphers:            x.Ciphers,
		Nonces:             x.Nonces,
		RoutingHeaderBytes: x.RoutingHeaderBytes,
	})
	// log.D.S("intros", ng.HiddenRouting)
	// log.D.S(ng.GetLocalNodeAddressString(), ng.HiddenRouting)
	log.D.Ln("stored new introduction, starting broadcast")
	go GossipIntro(&x.Intro, ng.SessionManager, ng.C)
	return
}

func MakeHiddenService(in *Intro, alice, bob *SessionData,
	c Circuit, ks *signer.KeySet) Skins {
	
	headers := GetHeaders(alice, bob, c, ks)
	return Skins{}.
		RoutingHeader(headers.Forward).
		HiddenService(in, headers.ExitPoint()).
		RoutingHeader(headers.Return)
}

func (ng *Engine) SendHiddenService(id nonce.ID, key *prv.Key,
	expiry time.Time, alice, bob *SessionData,
	svc *Service, hook Callback) (in *Intro) {
	
	hops := StandardCircuit()
	s := make(Sessions, len(hops))
	s[2] = alice
	se := ng.SelectHops(hops, s, "sendhiddenservice")
	var c Circuit
	copy(c[:], se[:len(c)])
	in = NewIntro(id, key, alice.Node.AddrPort, expiry)
	// log.D.S("intro", in, in.Validate())
	o := MakeHiddenService(in, alice, bob, c, ng.KeySet)
	// log.D.S("hidden service onion", o)
	log.D.F("%s sending out hidden service onion %s",
		ng.GetLocalNodeAddressString(),
		color.Yellow.Sprint(alice.Node.AddrPort.String()))
	res := ng.PostAcctOnion(o)
	// log.D.S("hs onion binary", res.B.ToBytes())
	ng.HiddenRouting.AddHiddenService(svc, key, in,
		ng.GetLocalNodeAddressString())
	// log.D.S("storing hidden service info", ng.HiddenRouting)
	ng.SendWithOneHook(c[0].Node.AddrPort, res, hook, ng.PendingResponses)
	return
}
