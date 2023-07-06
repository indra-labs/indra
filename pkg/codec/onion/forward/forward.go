// Package forward is an onion message layer that specifies a single redirection for the remainder of the onion.
//
// This is in contrast to the reverse message, which contains a 3 layer header, needed for when there will be a reply, or just for a compact and indistinguishable 1, 2 or 3 hop relaying that does not leak its position in the route.
package forward

import (
	"github.com/indra-labs/indra/pkg/codec"
	"github.com/indra-labs/indra/pkg/codec/onion/crypt"
	"github.com/indra-labs/indra/pkg/codec/onion/end"
	"github.com/indra-labs/indra/pkg/codec/ont"
	"github.com/indra-labs/indra/pkg/codec/reg"
	"github.com/indra-labs/indra/pkg/engine/magic"
	"github.com/indra-labs/indra/pkg/engine/sess"
	"github.com/indra-labs/indra/pkg/engine/sessions"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/splice"
	"net/netip"
	"reflect"
)

var (
	log   = log2.GetLogger()
	fails = log.E.Chk
)

const (
	Magic = "forw"
	Len   = magic.Len + 1 + splice.AddrLen
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
	AddrPort *netip.AddrPort
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
	if e = magic.TooShort(s.Remaining(), Len-magic.Len,
		Magic); fails(e) {
		return
	}
	s.ReadAddrPort(&x.AddrPort)
	return
}

// Encode a Forward into the next bytes of a splice.Splice.
func (x *Forward) Encode(s *splice.Splice) error {
	log.T.F("encoding %s %s", reflect.TypeOf(x),
		x.AddrPort.String(),
	)
	return x.Onion.Encode(s.Magic(Magic).AddrPort(x.AddrPort))
}

// GetOnion returns the onion inside this Forward.
func (x *Forward) GetOnion() interface{} { return x.Onion }

// Handle is the relay logic for an engine handling a Forward message.
func (x *Forward) Handle(s *splice.Splice, p ont.Onion, ng ont.Ngin) (e error) {
	// Forward the whole buffer received onwards. Usually there will be a
	// crypt.Layer under this which will be unwrapped by the receiver.
	if x.AddrPort.String() == ng.Mgr().GetLocalNodeAddress().String() {
		// it is for us, we want to unwrap the next part.
		ng.HandleMessage(splice.BudgeUp(s), x)
	} else {
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
		ng.Mgr().Send(x.AddrPort, splice.BudgeUp(s))
	}
	return e
}

// Len returns the length of this Forward message.
func (x *Forward) Len() int { return Len + x.Onion.Len() }

// Magic is the identifying 4 byte string indicating an Forward message follows.
func (x *Forward) Magic() string { return Magic }

// Wrap puts another onion inside this Forward onion.
func (x *Forward) Wrap(inner ont.Onion) { x.Onion = inner }

// New creates a new Forward onion.
func New(addr *netip.AddrPort) ont.Onion { return &Forward{AddrPort: addr, Onion: &end.End{}} }

// Gen is a factory function for a Forward.
func Gen() codec.Codec { return &Forward{} }

func init() { reg.Register(Magic, Gen) }
