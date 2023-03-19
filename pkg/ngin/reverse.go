package ngin

import (
	"net/netip"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/ciph"
	"git-indra.lan/indra-labs/indra/pkg/ngin/magic"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
	"git-indra.lan/indra-labs/indra/pkg/util/zip"
)

const (
	ReverseMagic = "rv"
	ReverseLen   = magic.Len + 1 + zip.AddrLen
)

type Reverse struct {
	*netip.AddrPort
	Onion
}

func reversePrototype() Onion { return &Reverse{} }

func init() { Register(ReverseMagic, reversePrototype) }

func (o Skins) Reverse(ip *netip.AddrPort) Skins {
	return append(o, &Reverse{AddrPort: ip, Onion: nop})
}

func (x *Reverse) Magic() string { return ReverseMagic }

func (x *Reverse) Encode(s *zip.Splice) error {
	return x.Onion.Encode(s.Magic(ReverseMagic).AddrPort(x.AddrPort))
}

func (x *Reverse) Decode(s *zip.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), ReverseLen-magic.Len,
		ReverseMagic); check(e) {
		return
	}
	s.ReadAddrPort(&x.AddrPort)
	return
}

func (x *Reverse) Len() int { return ReverseLen + x.Onion.Len() }

func (x *Reverse) Wrap(inner Onion) { x.Onion = inner }

func (x *Reverse) Handle(s *zip.Splice, p Onion,
	ng *Engine) (e error) {
	
	if x.AddrPort.String() == ng.GetLocalNodeAddress().String() {
		in := Recognise(s, ng.GetLocalNodeAddress())
		if e = in.Decode(s); check(e) {
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
		ciph.Encipher(ciph.GetBlock(on.ToPriv, on.FromPub), on.Nonce,
			s.GetRange(c, c+2*ReverseCryptLen))
		// shift the header segment upwards and pad the
		// remainder.
		s.CopyRanges(start, first, first, second)
		s.CopyRanges(first, second, second, last)
		s.CopyIntoRange(slice.NoisePad(ReverseCryptLen), second, last)
		if last != s.Len() {
			ciph.Encipher(ciph.GetBlock(pld, on.FromPub), on.Nonce,
				s.GetRange(last, -1))
		}
		if string(s.GetRange(start, start+magic.Len)) != ReverseMagic {
			// It's for us!
			log.T.S("handling response",
				// s.GetRange(last, -1).ToBytes(),
			)
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
