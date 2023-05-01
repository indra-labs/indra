package onions

import (
	"reflect"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/engine/coding"
	"git-indra.lan/indra-labs/indra/pkg/engine/magic"
	"git-indra.lan/indra-labs/indra/pkg/engine/sess"
	"git-indra.lan/indra-labs/indra/pkg/engine/sessions"
	"git-indra.lan/indra-labs/indra/pkg/util/splice"
)

const (
	HiddenServiceMagic = "hids"
	HiddenServiceLen   = magic.Len + IntroLen +
		3*sha256.Len + nonce.IVLen*3 + RoutingHeaderLen
)

type HiddenService struct {
	Intro
	// Ciphers is a set of 3 symmetric ciphers that are to be used in their
	// given order over the reply message from the service.
	crypto.Ciphers
	// Nonces are the nonces to use with the cipher when creating the encryption
	// for the reply message, they are common with the crypts in the header.
	crypto.Nonces
	RoutingHeaderBytes
	Onion
}

func hiddenServiceGen() coding.Codec           { return &HiddenService{} }
func init()                                    { Register(HiddenServiceMagic, hiddenServiceGen) }
func (x *HiddenService) Magic() string         { return HiddenServiceMagic }
func (x *HiddenService) Len() int              { return HiddenServiceLen + x.Onion.Len() }
func (x *HiddenService) Wrap(inner Onion)      { x.Onion = inner }
func (x *HiddenService) GetOnion() interface{} { return x }

func (x *HiddenService) Encode(s *splice.Splice) (e error) {
	log.T.S("encoding", reflect.TypeOf(x),
		x.ID, x.Key, x.AddrPort, x.Ciphers, x.Nonces, x.RoutingHeaderBytes,
	)
	SpliceIntro(s.Magic(HiddenServiceMagic), &x.Intro)
	return x.Onion.Encode(s.Ciphers(x.Ciphers).Nonces(x.Nonces))
}

func (x *HiddenService) Decode(s *splice.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), HiddenServiceLen-magic.Len,
		HiddenServiceMagic); fails(e) {
		return
	}
	if e = x.Intro.Decode(s); fails(e) {
		return
	}
	s.
		ReadCiphers(&x.Ciphers).
		ReadNonces(&x.Nonces)
	rb := GetRoutingHeaderFromCursor(s)
	ReadRoutingHeader(s, &rb)
	return
}

func (x *HiddenService) Handle(s *splice.Splice, p Onion, ng Ngin) (e error) {
	log.D.F("%s adding introduction for key %s",
		ng.Mgr().GetLocalNodeAddressString(), x.Key.ToBase32Abbreviated())
	ng.GetHidden().AddIntro(x.Key, &Introduction{
		Intro: &x.Intro,
		ReplyHeader: ReplyHeader{
			Ciphers:            x.Ciphers,
			Nonces:             x.Nonces,
			RoutingHeaderBytes: x.RoutingHeaderBytes,
		},
	})
	log.D.Ln("stored new introduction, starting broadcast")
	go GossipIntro(&x.Intro, ng.Mgr(), ng.KillSwitch())
	return
}

func (x *HiddenService) Account(res *sess.Data, sm *sess.Manager, s *sessions.Data, last bool) (skip bool, sd *sessions.Data) {
	
	res.ID = x.Intro.ID
	res.Billable = append(res.Billable, s.ID)
	skip = true
	return
}
