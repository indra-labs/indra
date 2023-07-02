// Package hiddenservice provides a message type for hidden services to send to designated introducer relays.
//
// These are a message type that does not go into a peer's ad bundle, they are simply gossiped when they are received, and will stop being gossiped after the expiry in the embedded intro message expiry is passed.
//
// Of course a hidden service can decide to unilaterally stop sending these before this expiry for any reason, and each new client triggers the generation of a new intro, which is forwarded to the introducer after it performs an introduction with this data, but usually it will continue to give a new intro each time a client connects through until the expiry time.
package hiddenservice

import (
	"github.com/indra-labs/indra/pkg/hidden"
	intro "github.com/indra-labs/indra/pkg/onions/adintro"
	"github.com/indra-labs/indra/pkg/onions/consts"
	"github.com/indra-labs/indra/pkg/onions/end"
	"github.com/indra-labs/indra/pkg/onions/exit"
	"github.com/indra-labs/indra/pkg/onions/ont"
	"github.com/indra-labs/indra/pkg/onions/reg"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
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

var (
	log   = log2.GetLogger()
	fails = log.E.Chk
)

const (
	HiddenServiceMagic = "hids"
	HiddenServiceLen   = magic.Len +
		intro.Len +
		3*sha256.Len +
		nonce.IVLen*3 +
		consts.RoutingHeaderLen
)

type HiddenService struct {
	Intro intro.Ad
	// Ciphers is a set of 3 symmetric ciphers that are to be used in their
	// given order over the reply message from the service.
	crypto.Ciphers
	// Nonces are the nonces to use with the cipher when creating the encryption
	// for the reply message, they are common with the crypts in the header.
	crypto.Nonces
	hidden.RoutingHeaderBytes
	ont.Onion
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
	rb := hidden.GetRoutingHeaderFromCursor(s)
	hidden.ReadRoutingHeader(s, &rb)
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

func (x *HiddenService) Handle(s *splice.Splice, p ont.Onion, ng ont.Ngin) (e error) {
	log.D.S("intro", x.Intro)
	log.D.F("%s adding introduction for key %s",
		ng.Mgr().GetLocalNodeAddressString(),
		x.Intro.
			Key)
	ng.GetHidden().AddIntro(x.Intro.Key, &hidden.Introduction{
		Intro: &x.Intro,
		ReplyHeader: hidden.ReplyHeader{
			Ciphers:            x.Ciphers,
			Nonces:             x.Nonces,
			RoutingHeaderBytes: x.RoutingHeaderBytes,
		},
	})
	log.D.Ln("stored new introduction, starting broadcast")
	// go x.Intro.Gossip(ng.Mgr(), ng.WaitForShutdown())
	return
}

func (x *HiddenService) Len() int             { return HiddenServiceLen + x.Onion.Len() }
func (x *HiddenService) Magic() string        { return HiddenServiceMagic }
func (x *HiddenService) Wrap(inner ont.Onion) { x.Onion = inner }

func NewHiddenService(in *intro.Ad, point *exit.ExitPoint) ont.Onion {
	return &HiddenService{
		Intro:   *in,
		Ciphers: crypto.GenCiphers(point.Keys, point.ReturnPubs),
		Nonces:  point.Nonces,
		Onion:   end.NewEnd(),
	}
}
func hiddenServiceGen() coding.Codec { return &HiddenService{} }
func init()                          { reg.Register(HiddenServiceMagic, hiddenServiceGen) }
