package engine

import (
	"net/netip"
	
	"git-indra.lan/indra-labs/indra/pkg/engine/magic"
)

const (
	ForwardMagic = "fw"
	ForwardLen   = magic.Len + 1 + AddrLen
)

type Forward struct {
	AddrPort *netip.AddrPort
	Onion
}

func forwardPrototype() Onion { return &Forward{} }

func init() { Register(ForwardMagic, forwardPrototype) }

func (o Skins) Forward(addr *netip.AddrPort) Skins {
	return append(o, &Forward{AddrPort: addr, Onion: &End{}})
}

func (x *Forward) Magic() string { return ForwardMagic }

func (x *Forward) Encode(s *Splice) error {
	// log.T.S("encoding", reflect.TypeOf(x),
	// 	x.AddrPort.String(),
	// )
	return x.Onion.Encode(s.Magic(ForwardMagic).AddrPort(x.AddrPort))
}

func (x *Forward) Decode(s *Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), ForwardLen-magic.Len,
		ForwardMagic); check(e) {
		return
	}
	s.ReadAddrPort(&x.AddrPort)
	return
}

func (x *Forward) Len() int { return ForwardLen + x.Onion.Len() }

func (x *Forward) Wrap(inner Onion) { x.Onion = inner }

func (x *Forward) Handle(s *Splice, p Onion,
	ng *Engine) (e error) {
	
	// Forward the whole buffer received onwards. Usually there will be a
	// crypt.Layer under this which will be unwrapped by the receiver.
	if x.AddrPort.String() == ng.GetLocalNodeAddressString() {
		// it is for us, we want to unwrap the next part.
		ng.HandleMessage(BudgeUp(s), x)
	} else {
		switch on1 := p.(type) {
		case *Crypt:
			sess := ng.FindSessionByHeader(on1.ToPriv)
			if sess != nil {
				ng.DecSession(sess.ID,
					ng.GetLocalNodeRelayRate()*s.Len(),
					false, "forward")
			}
		}
		// we need to forward this message onion.
		ng.Send(x.AddrPort, BudgeUp(s))
	}
	return e
}
