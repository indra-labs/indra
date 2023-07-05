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
	ForwardMagic = "forw"
	ForwardLen   = magic.Len + 1 + splice.AddrLen
)

type Forward struct {
	AddrPort *netip.AddrPort
	ont.Onion
}

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

func (x *Forward) Decode(s *splice.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), ForwardLen-magic.Len,
		ForwardMagic); fails(e) {
		return
	}
	s.ReadAddrPort(&x.AddrPort)
	return
}

func (x *Forward) Encode(s *splice.Splice) error {
	log.T.F("encoding %s %s", reflect.TypeOf(x),
		x.AddrPort.String(),
	)
	return x.Onion.Encode(s.Magic(ForwardMagic).AddrPort(x.AddrPort))
}

func (x *Forward) GetOnion() interface{} { return x }

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

func (x *Forward) Len() int                     { return ForwardLen + x.Onion.Len() }
func (x *Forward) Magic() string                { return ForwardMagic }
func (x *Forward) Wrap(inner ont.Onion)         { x.Onion = inner }
func NewForward(addr *netip.AddrPort) ont.Onion { return &Forward{AddrPort: addr, Onion: &end.End{}} }
func forwardGen() codec.Codec                   { return &Forward{} }
func init()                                     { reg.Register(ForwardMagic, forwardGen) }
