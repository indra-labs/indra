// Package hiddenservice provides a message type for hidden services to send to designated introducer relays.
//
// These are a message type that does not go into a peer's ad bundle, they are simply gossiped when they are received, and will stop being gossiped after the expiry in the embedded intro message expiry is passed.
//
// Of course a hidden service can decide to unilaterally stop sending these before this expiry for any reason, and each new client triggers the generation of a new intro, which is forwarded to the introducer after it performs an introduction with this data, but usually it will continue to give a new intro each time a client connects through until the expiry time.
package hiddenservice

import (
	"github.com/indra-labs/indra/pkg/codec"
	"github.com/indra-labs/indra/pkg/codec/ad/intro"
	"github.com/indra-labs/indra/pkg/codec/onion/end"
	"github.com/indra-labs/indra/pkg/codec/onion/exit"
	"github.com/indra-labs/indra/pkg/codec/ont"
	"github.com/indra-labs/indra/pkg/codec/reg"
	"github.com/indra-labs/indra/pkg/engine/consts"
	"github.com/indra-labs/indra/pkg/hidden"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"reflect"

	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/crypto/sha256"
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
	Magic = "hids"
	Len   = magic.Len +
		intro.Len +
		3*sha256.Len +
		nonce.IVLen*3 +
		consts.RoutingHeaderLen
)

// HiddenService is a message providing an Intro and the necessary Ciphers,
// Nonces and RoutingHeaderBytes to forward a Route message through to the hidden
// service, using the client's reply RoutingHeader.
type HiddenService struct {

	// Intro of the hidden service.
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

// Account the traffic for the relay handling this message.
func (x *HiddenService) Account(res *sess.Data, sm *sess.Manager, s *sessions.Data,
	last bool) (skip bool, sd *sessions.Data) {

	res.ID = x.Intro.ID
	res.Billable = append(res.Billable, s.Header.Bytes)
	skip = true
	return
}

// Decode the HiddenService message from the next bytes of a Splice.
func (x *HiddenService) Decode(s *splice.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), Len-magic.Len, Magic); fails(e) {
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

// Encode a HiddenService into a the next bytes of a Splice.
func (x *HiddenService) Encode(s *splice.Splice) (e error) {
	log.T.S("encoding", reflect.TypeOf(x),
		x.Intro.ID, x.Intro.Key, x.Intro.Introducer, x.Ciphers, x.Nonces, x.RoutingHeaderBytes,
	)
	x.Intro.Splice(s.Magic(Magic))
	return x.Onion.Encode(s.Ciphers(x.Ciphers).Nonces(x.Nonces))
}

// GetOnion returns the inner onion or remaining parts of the message prototype.
func (x *HiddenService) GetOnion() interface{} { return x }

// Handle defines how the ont.Ngin should deal with this onion type.
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

// Len returns the length of the onion starting from this one (used to size a
// Splice).
func (x *HiddenService) Len() int { return Len + x.Onion.Len() }

// Magic bytes identifying a HiddenService message is up next.
func (x *HiddenService) Magic() string { return Magic }

// Wrap places another onion inside this one in its slot.
func (x *HiddenService) Wrap(inner ont.Onion) { x.Onion = inner }

// NewHiddenService generates a new HiddenService data structure and returns it
// as an ont.Onion interface.
func NewHiddenService(in *intro.Ad, point *exit.ExitPoint) ont.Onion {
	return &HiddenService{
		Intro:   *in,
		Ciphers: crypto.GenCiphers(point.Keys, point.ReturnPubs),
		Nonces:  point.Nonces,
		Onion:   end.NewEnd(),
	}
}

// Gen is a factory function for a HiddenService.
func Gen() codec.Codec { return &HiddenService{} }

func init() { reg.Register(Magic, Gen) }
