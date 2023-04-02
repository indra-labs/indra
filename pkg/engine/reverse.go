package engine

import (
	"net/netip"
	"reflect"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/ciph"
	"git-indra.lan/indra-labs/indra/pkg/engine/magic"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

const (
	ReverseMagic = "rv"
	ReverseLen   = magic.Len + 1 + AddrLen
)

type Reverse struct {
	AddrPort *netip.AddrPort
	Onion
}

func reversePrototype() Onion       { return &Reverse{} }
func init()                         { Register(ReverseMagic, reversePrototype) }
func (x *Reverse) Magic() string    { return ReverseMagic }
func (x *Reverse) Len() int         { return ReverseLen + x.Onion.Len() }
func (x *Reverse) Wrap(inner Onion) { x.Onion = inner }

func (o Skins) Reverse(ip *netip.AddrPort) Skins {
	return append(o, &Reverse{AddrPort: ip, Onion: nop})
}

func (x *Reverse) Encode(s *Splice) (e error) {
	log.T.Ln("encoding", reflect.TypeOf(x), x.AddrPort)
	if x.AddrPort == nil {
		s.Advance(ReverseLen, "reverse")
	} else {
		s.Magic(ReverseMagic).AddrPort(x.AddrPort)
	}
	if x.Onion != nil {
		e = x.Onion.Encode(s)
	}
	return
}

func (x *Reverse) Decode(s *Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), ReverseLen-magic.Len,
		ReverseMagic); fails(e) {
		return
	}
	s.ReadAddrPort(&x.AddrPort)
	return
}

func (x *Reverse) Handle(s *Splice, p Onion,
	ng *Engine) (e error) {
	
	if x.AddrPort.String() == ng.GetLocalNodeAddress().String() {
		in := Recognise(s, ng.GetLocalNodeAddress())
		if e = in.Decode(s); fails(e) {
			return e
		}
		if in.Magic() != CryptMagic {
			return e
		}
		on := in.(*Crypt)
		first := s.GetCursor()
		start := first - ReverseCryptLen
		second := first + ReverseCryptLen
		last := second + ReverseCryptLen
		hdr, pld, _, _ := ng.FindCloaked(on.Cloak)
		if hdr == nil || pld == nil {
			log.E.F("failed to find key for %s",
				ng.GetLocalNodeAddressString())
			return e
		}
		// We need to find the PayloadPub to match.
		on.ToPriv = hdr
		// Decrypt using the Payload key and header nonce.
		c := s.GetCursor()
		ciph.Encipher(ciph.GetBlock(on.ToPriv, on.FromPub, "reverse header"),
			on.Nonce, s.GetRange(c, c+2*ReverseCryptLen))
		// shift the header segment upwards and pad the
		// remainder.
		s.CopyRanges(start, first, first, second)
		s.CopyRanges(first, second, second, last)
		s.CopyIntoRange(slice.NoisePad(ReverseCryptLen), second, last)
		if last != s.Len() {
			ciph.Encipher(ciph.GetBlock(pld, on.FromPub, "reverse payload"),
				on.Nonce, s.GetFrom(last))
		}
		if string(s.GetRange(start, start+magic.Len)) != ReverseMagic {
			// It's for us!
			log.T.S("handling response")
			ng.HandleMessage(BudgeUp(s.SetCursor(last)), on)
			return e
		}
		sess := ng.FindSessionByHeader(hdr)
		if sess != nil {
			ng.DecSession(sess.ID,
				ng.GetLocalNodeRelayRate()*s.Len(), false, "reverse")
			ng.HandleMessage(BudgeUp(s.SetCursor(start)), on)
		}
	} else if p != nil {
		// we need to forward this message onion.
		log.T.Ln(ng.GetLocalNodeAddressString(), "forwarding reverse")
		ng.Send(x.AddrPort, s)
	} else {
		log.E.Ln("we do not forward nonsense! scoff! snort!")
	}
	return e
}

func (x *Reverse) Account(res *SendData, sm *SessionManager,
	s *SessionData, last bool) (skip bool, sd *SessionData) {
	
	res.Billable = append(res.Billable, s.ID)
	return
}
