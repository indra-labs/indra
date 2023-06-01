package onions

import (
	"github.com/indra-labs/indra/pkg/onions/intro"
	"github.com/indra-labs/indra/pkg/onions/reg"
	"reflect"

	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/crypto/sha256"
	"github.com/indra-labs/indra/pkg/engine/coding"
	"github.com/indra-labs/indra/pkg/engine/magic"
	"github.com/indra-labs/indra/pkg/engine/sess"
	"github.com/indra-labs/indra/pkg/engine/sessions"
	"github.com/indra-labs/indra/pkg/util/splice"
)

const (
	HiddenServiceMagic = "hids"
	HiddenServiceLen   = magic.Len + intro.Len +
		3*sha256.Len + nonce.IVLen*3 + RoutingHeaderLen
)

type HiddenService struct {
	Intro intro.Ad
	// Ciphers is a set of 3 symmetric ciphers that are to be used in their
	// given order over the reply message from the service.
	crypto.Ciphers
	// Nonces are the nonces to use with the cipher when creating the encryption
	// for the reply message, they are common with the crypts in the header.
	crypto.Nonces
	RoutingHeaderBytes
	Onion
}

func (x *HiddenService) Account(res *sess.Data, sm *sess.Manager, s *sessions.Data, last bool) (skip bool,
	sd *sessions.Data) {

	res.ID = x.Intro.ID
	res.Billable = append(res.Billable, s.Header.Bytes)
	skip = true
	return
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

func (x *HiddenService) Encode(s *splice.Splice) (e error) {
	log.T.S("encoding", reflect.TypeOf(x),
		x.Intro.ID, x.Intro.Key, x.Intro.AddrPort, x.Ciphers, x.Nonces, x.RoutingHeaderBytes,
	)
	x.Intro.Splice(s.Magic(HiddenServiceMagic))
	return x.Onion.Encode(s.Ciphers(x.Ciphers).Nonces(x.Nonces))
}

func (x *HiddenService) GetOnion() interface{} { return x }

func (x *HiddenService) Handle(s *splice.Splice, p Onion, ng Ngin) (e error) {
	log.D.F("%s adding introduction for key %s",
		ng.Mgr().GetLocalNodeAddressString(), x.Intro.Key.ToBased32Abbreviated())
	ng.GetHidden().AddIntro(x.Intro.Key, &Introduction{
		Intro: &x.Intro,
		ReplyHeader: ReplyHeader{
			Ciphers:            x.Ciphers,
			Nonces:             x.Nonces,
			RoutingHeaderBytes: x.RoutingHeaderBytes,
		},
	})
	log.D.Ln("stored new introduction, starting broadcast")
	go x.Intro.Gossip(ng.Mgr(), ng.KillSwitch())
	return
}

func (x *HiddenService) Len() int         { return HiddenServiceLen + x.Onion.Len() }
func (x *HiddenService) Magic() string    { return HiddenServiceMagic }
func (x *HiddenService) Wrap(inner Onion) { x.Onion = inner }

func NewHiddenService(in *intro.Ad, point *ExitPoint) Onion {
	return &HiddenService{
		Intro:   *in,
		Ciphers: crypto.GenCiphers(point.Keys, point.ReturnPubs),
		Nonces:  point.Nonces,
		Onion:   NewEnd(),
	}
}
func hiddenServiceGen() coding.Codec { return &HiddenService{} }
func init()                          { reg.Register(HiddenServiceMagic, hiddenServiceGen) }
