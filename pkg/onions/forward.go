package onions

import (
	"github.com/indra-labs/indra/pkg/engine/coding"
	"github.com/indra-labs/indra/pkg/engine/magic"
	"github.com/indra-labs/indra/pkg/engine/sess"
	"github.com/indra-labs/indra/pkg/engine/sessions"
	"github.com/indra-labs/indra/pkg/onions/reg"
	"github.com/indra-labs/indra/pkg/util/splice"
	"net/netip"
	"reflect"
)

const (
	ForwardMagic = "forw"
	ForwardLen   = magic.Len + 1 + splice.AddrLen
)

type Forward struct {
	AddrPort *netip.AddrPort
	Onion
}

func (x *Forward) Account(res *sess.Data, sm *sess.Manager,
	s *sessions.Data, last bool) (skip bool, sd *sessions.Data) {
	res.Billable = append(res.Billable, s.Header.Bytes)
	res.PostAcct = append(res.PostAcct,
		func() {
			sm.DecSession(s.Header.Bytes, s.Node.RelayRate*len(res.B),
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

func (x *Forward) Handle(s *splice.Splice, p Onion, ng Ngin) (e error) {
	// Forward the whole buffer received onwards. Usually there will be a
	// crypt.Layer under this which will be unwrapped by the receiver.
	if x.AddrPort.String() == ng.Mgr().GetLocalNodeAddress().String() {
		// it is for us, we want to unwrap the next part.
		ng.HandleMessage(splice.BudgeUp(s), x)
	} else {
		switch on1 := p.(type) {
		case *Crypt:
			sess := ng.Mgr().FindSessionByHeader(on1.ToPriv)
			if sess != nil {
				ng.Mgr().DecSession(sess.Header.Bytes,
					ng.Mgr().GetLocalNodeRelayRate()*s.Len(),
					false, "forward")
			}
		}
		// we need to forward this message onion.
		ng.Mgr().Send(x.AddrPort, splice.BudgeUp(s))
	}
	return e
}

func (x *Forward) Len() int         { return ForwardLen + x.Onion.Len() }
func (x *Forward) Magic() string    { return ForwardMagic }
func (x *Forward) Wrap(inner Onion) { x.Onion = inner }
func forwardGen() coding.Codec      { return &Forward{} }
func init()                         { reg.Register(ForwardMagic, forwardGen) }
