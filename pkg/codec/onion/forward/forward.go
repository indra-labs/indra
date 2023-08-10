// Package forward is an onion message layer that specifies a single redirection for the remainder of the onion.
//
// This is in contrast to the reverse message, which contains a 3 layer header, needed for when there will be a reply, or just for a compact and indistinguishable 1, 2 or 3 hop relaying that does not leak its position in the route.
package forward

import (
	"git.indra-labs.org/dev/ind/pkg/cfg"
	"git.indra-labs.org/dev/ind/pkg/codec"
	"git.indra-labs.org/dev/ind/pkg/codec/onion/cores/end"
	"git.indra-labs.org/dev/ind/pkg/codec/onion/crypt"
	"git.indra-labs.org/dev/ind/pkg/codec/ont"
	"git.indra-labs.org/dev/ind/pkg/codec/reg"
	"git.indra-labs.org/dev/ind/pkg/engine/magic"
	"git.indra-labs.org/dev/ind/pkg/engine/sess"
	"git.indra-labs.org/dev/ind/pkg/engine/sessions"
	log2 "git.indra-labs.org/dev/ind/pkg/proc/log"
	"git.indra-labs.org/dev/ind/pkg/util/splice"
	"github.com/multiformats/go-multiaddr"
	"reflect"
)

var (
	log   = log2.GetLogger()
	fails = log.E.Chk
)

const (
	Magic = "forw"
)

// Forward is a simple forward of the remainder of the onion to a given address
// and port. Note that we don't use the key, just the address. Relays can have
// multiple addresses but for a given message, one of them is chosen.
//
// If a reply is required, the reverse.Reverse is used with a RoutingHeader.
//
// todo: currently clients expect that the different addresses are tunnels to the
//
//	same relay. They are considered to be the same physical connection with an
//	extension and are weighted equally. Perhaps they should have bandwidth
//	capacity indications?
type Forward struct {
	Multiaddr multiaddr.Multiaddr
	ont.Onion
}

// Account for the Forward message is straightforward, message size/Mb/RelayRate.
func (x *Forward) Account(res *sess.Data, sm *sess.Manager,
	s *sessions.Data, last bool) (skip bool, sd *sessions.Data) {
	res.Billable = append(res.Billable, s.Header.Bytes)
	res.PostAcct = append(res.PostAcct,
		func() {
			sm.DecSession(s.Header.Bytes, int(s.Node.RelayRate)*len(res.B),
				true, "forward")
		})
	return
}

// Decode a Forward from a provided splice.Splice.
func (x *Forward) Decode(s *splice.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), magic.Len,
		Magic); fails(e) {
		return
	}
	s.ReadMultiaddr(&x.Multiaddr)
	return
}

// Encode a Forward into the next bytes of a splice.Splice.
func (x *Forward) Encode(s *splice.Splice) error {
	log.T.F("encoding %s %s", reflect.TypeOf(x),
		x.Multiaddr.String(),
	)
	return x.Onion.Encode(s.Magic(Magic).
		Multiaddr(x.Multiaddr, cfg.GetCurrentDefaultPort()))
}

// Unwrap returns the onion inside this Forward.
func (x *Forward) Unwrap() interface{} { return x.Onion }

// Handle is the relay logic for an engine handling a Forward message.
func (x *Forward) Handle(s *splice.Splice, p ont.Onion, ng ont.Ngin) (e error) {

	// Forward the whole buffer received onwards.
	//
	// Usually there will be a crypt.Layer under this which will be unwrapped by
	// the receiver.
	if ng.Mgr().MatchesLocalNodeAddress(x.Multiaddr) {

		// it is for us, we want to unwrap the next part.
		ng.HandleMessage(splice.BudgeUp(s), x)
	} else {
		// accounting on the relevant session and then forward the message.
		switch on1 := p.(type) {
		case *crypt.Crypt:
			sess := ng.Mgr().FindSessionByHeader(on1.ToPriv)
			if sess != nil {
				ng.Mgr().DecSession(sess.Header.Bytes,
					int(ng.Mgr().GetLocalNodeRelayRate())*s.Len(),
					false, "forward")
			}
		}
		// we need to forward this message onion.
		ng.Mgr().Send(x.Multiaddr, splice.BudgeUp(s))
	}
	return e
}

// Len returns the length of this Forward message.
func (x *Forward) Len() int {

	codec.MustNotBeNil(x)

	// b, _ := multi.AddrToBytes(x.Multiaddr,
	// 	cfg.GetCurrentDefaultPort())
	return magic.Len + 21 + x.Onion.Len()
}

// Magic is the identifying 4 byte string indicating an Forward message follows.
func (x *Forward) Magic() string { return Magic }

// Wrap puts another onion inside this Forward onion.
func (x *Forward) Wrap(inner ont.Onion) { x.Onion = inner }

// New creates a new Forward onion.
func New(addr multiaddr.Multiaddr) ont.Onion {
	return &Forward{Multiaddr: addr, Onion: &end.End{}}
}

// Gen is a factory function for a Forward.
func Gen() codec.Codec { return &Forward{} }

func init() { reg.Register(Magic, Gen) }
