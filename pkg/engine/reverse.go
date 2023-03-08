package engine

import (
	"net/netip"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/ciph"
	"git-indra.lan/indra-labs/indra/pkg/util/octet"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

const (
	ReverseMagic = "rv"
	ReverseLen   = MagicLen + 1 + octet.AddrLen
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

func (x *Reverse) Encode(s *octet.Splice) error {
	return x.Onion.Encode(s.Magic(ReverseMagic).AddrPort(x.AddrPort))
}

func (x *Reverse) Decode(s *octet.Splice) (e error) {
	if e = TooShort(s.Remaining(), ReverseLen-MagicLen,
		ReverseMagic); check(e) {
		return
	}
	s.ReadAddrPort(&x.AddrPort)
	return
}

func (x *Reverse) Len() int { return ReverseLen + x.Onion.Len() }

func (x *Reverse) Wrap(inner Onion) { x.Onion = inner }

func (x *Reverse) Handle(s *octet.Splice, p Onion,
	ng *Engine) (e error) {
	
	if x.AddrPort.String() == ng.GetLocalNodeAddress().String() {
		in := Recognise(s)
		if e = in.Decode(s); check(e) {
			return e
		}
		if in.Magic() != CryptMagic {
			return e
		}
		on := in.(*Crypt)
		first := s.GetCursor()
		log.D.S("reverse segments",
			s.GetRange(-1, first).ToBytes(),
			s.GetRange(first, -1).ToBytes(),
			// s.GetRange(-1, -1).ToBytes(),
		)
		start := first - ReverseLayerLen
		second := first + ReverseLayerLen
		last := second + ReverseLayerLen
		log.T.Ln("searching for reverse crypt keys")
		hdr, pld, _, _ := ng.FindCloaked(on.Cloak)
		if hdr == nil || pld == nil {
			log.E.F("failed to find key for %s",
				ng.GetLocalNodeAddress().String())
			return e
		}
		// We need to find the PayloadPub to match.
		on.ToPriv = hdr
		// Decrypt using the Payload key and header nonce.
		c := s.GetCursor()
		log.T.S("segments",
			s.GetRange(-1, start).ToBytes(),
			s.GetRange(start, first).ToBytes(),
			s.GetRange(first, second).ToBytes(),
			s.GetRange(second, last).ToBytes(),
			s.GetRange(last, -1).ToBytes(),
		)
		ciph.Encipher(ciph.GetBlock(on.ToPriv, on.FromPub), on.Nonce,
			s.GetRange(c, c+2*ReverseLayerLen))
		log.T.S("header decoded",
			s.GetRange(-1, start).ToBytes(),
			s.GetRange(start, first).ToBytes(),
			s.GetRange(first, second).ToBytes(),
			s.GetRange(second, last).ToBytes(),
			s.GetRange(last, -1).ToBytes(),
		)
		// shift the header segment upwards and pad the
		// remainder.
		s.CopyRanges(start, first, first, second)
		s.CopyRanges(first, second, second, last)
		s.CopyIntoRange(slice.NoisePad(ReverseLayerLen), second, last)
		log.T.S("header budged up",
			s.GetRange(-1, start).ToBytes(),
			s.GetRange(start, first).ToBytes(),
			s.GetRange(first, second).ToBytes(),
			s.GetRange(second, last).ToBytes(),
			s.GetRange(last, -1).ToBytes(),
		)
		if last != s.Len() {
			ciph.Encipher(ciph.GetBlock(pld, on.FromPub), on.Nonce,
				s.GetRange(last, -1))
			log.T.S("payload decrypted",
				s.GetRange(-1, start).ToBytes(),
				s.GetRange(start, first).ToBytes(),
				s.GetRange(first, second).ToBytes(),
				s.GetRange(second, last).ToBytes(),
				s.GetRange(last, -1).ToBytes(),
			)
		}
		if s.GetRange(start, start+MagicLen).String() != ReverseMagic {
			// It's for us!
			log.T.Ln("handling response")
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
		log.T.Ln("forwarding reverse")
		ng.Send(x.AddrPort, s)
	} else {
		log.E.Ln("we do not forward nonsense! scoff! snort!")
	}
	return e
}
