// Package introquery is an onion message that verifies a relay is an introducer for a given hidden service, returning its intro.Ad.
//
// After receiving this message if the intro is valid a client can use a route message to start a connection.
package introquery

import (
	"github.com/indra-labs/indra/pkg/codec"
	"github.com/indra-labs/indra/pkg/codec/ad/intro"
	"github.com/indra-labs/indra/pkg/codec/onion/crypt"
	"github.com/indra-labs/indra/pkg/codec/onion/end"
	"github.com/indra-labs/indra/pkg/codec/onion/exit"
	"github.com/indra-labs/indra/pkg/codec/ont"
	"github.com/indra-labs/indra/pkg/codec/reg"
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/crypto/sha256"
	"github.com/indra-labs/indra/pkg/engine/magic"
	"github.com/indra-labs/indra/pkg/engine/sess"
	"github.com/indra-labs/indra/pkg/engine/sessions"
	"github.com/indra-labs/indra/pkg/hidden"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/splice"
	"reflect"
)

var (
	log   = log2.GetLogger()
	fails = log.E.Chk
)

const (
	Magic = "intq"
	Len   = magic.Len + nonce.IDLen + crypto.PubKeyLen +
		3*sha256.Len + nonce.IVLen*3
)

// IntroQuery is a query message to return the Intro for a given hidden service
// key.
type IntroQuery struct {
	ID nonce.ID

	// Ciphers is a set of 3 symmetric ciphers that are to be used in their
	// given order over the reply message from the service.
	crypto.Ciphers

	// Nonces are the nonces to use with the cipher when creating the
	// encryption for the reply message,
	// they are common with the crypts in the header.
	crypto.Nonces

	// Key is the public key of the hidden service.
	Key *crypto.Pub

	// Onion contained in Introquery should be a RoutingHeader and cipher/nonce set.
	ont.Onion
}

// Account for an Introquery. In this case just the bytes size.
func (x *IntroQuery) Account(res *sess.Data, sm *sess.Manager,
	s *sessions.Data, last bool) (skip bool, sd *sessions.Data) {

	res.ID = x.ID
	res.Billable = append(res.Billable, s.Header.Bytes)
	skip = true
	return
}

// Decode what should be an IntroQuery message from a splice.Splice.
func (x *IntroQuery) Decode(s *splice.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), Len-magic.Len,
		Magic); fails(e) {
		return
	}
	s.ReadID(&x.ID).ReadCiphers(&x.Ciphers).ReadNonces(&x.Nonces).
		ReadPubkey(&x.Key)
	return
}

// Encode this IntroQuery into a splice.Splice's next bytes.
func (x *IntroQuery) Encode(s *splice.Splice) (e error) {
	log.T.S("encoding", reflect.TypeOf(x),
		x.ID, x.Key, x.Ciphers, x.Nonces,
	)
	return x.Onion.Encode(s.
		Magic(Magic).
		ID(x.ID).Ciphers(x.Ciphers).Nonces(x.Nonces).
		Pubkey(x.Key),
	)
}

// Unwrap returns the onion inside this IntroQuery message.
func (x *IntroQuery) Unwrap() interface{} { return x.Onion }

// Handle provides the relay switching logic for an engine handling an Introquery
// message.
func (x *IntroQuery) Handle(s *splice.Splice, p ont.Onion, ng ont.Ngin) (e error) {
	ng.GetHidden().Lock()
	log.D.Ln(ng.Mgr().GetLocalNodeAddressString(), "handling introquery", x.ID,
		x.Key.ToBased32Abbreviated())
	var ok bool
	var il *intro.Ad
	if il, ok = ng.GetHidden().KnownIntros[x.Key.ToBytes()]; !ok {
		// if the reply is zeroes the querant knows it needs to retry at a different
		// relay
		il = &intro.Ad{}
		ng.GetHidden().Unlock()
		log.E.Ln("intro not known")
		return
	}
	ng.GetHidden().Unlock()
	e = il.Encode(s)
	rb := hidden.FormatReply(hidden.GetRoutingHeaderFromCursor(s), x.Ciphers, x.Nonces,
		s.GetAll())
	switch on1 := p.(type) {
	case *crypt.Crypt:
		sess := ng.Mgr().FindSessionByHeader(on1.ToPriv)
		if sess != nil {
			in := int(sess.Node.RelayRate) * s.Len() / 2
			out := int(sess.Node.RelayRate) * rb.Len() / 2
			ng.Mgr().DecSession(sess.Header.Bytes, in+out, false, "introquery")
		}
	}
	ng.HandleMessage(rb, x)
	return
}

// Len returns the length of the onion starting from this one (used to size a
// Splice).
func (x *IntroQuery) Len() int { return Len + x.Onion.Len() }

// Magic bytes identifying a HiddenService message is up next.
func (x *IntroQuery) Magic() string { return Magic }

// Wrap places another onion inside this one in its slot.
func (x *IntroQuery) Wrap(inner ont.Onion) { x.Onion = inner }

// New generates a new IntroQuery data structure and returns it as an ont.Onion
// interface.
func New(id nonce.ID, hsk *crypto.Pub, exit *exit.ExitPoint) ont.Onion {
	return &IntroQuery{
		ID:      id,
		Ciphers: crypto.GenCiphers(exit.Keys, exit.ReturnPubs),
		Nonces:  exit.Nonces,
		Key:     hsk,
		Onion:   end.NewEnd(),
	}
}

// Gen is a factory function for an IntroQuery.
func Gen() codec.Codec { return &IntroQuery{} }

func init() { reg.Register(Magic, Gen) }
