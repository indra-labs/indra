package onions

import (
	"github.com/indra-labs/indra/pkg/crypto/ciph"
	"github.com/indra-labs/indra/pkg/engine/coding"
	"github.com/indra-labs/indra/pkg/engine/magic"
	"github.com/indra-labs/indra/pkg/engine/sess"
	"github.com/indra-labs/indra/pkg/engine/sessions"
	"github.com/indra-labs/indra/pkg/onions/reg"
	"github.com/indra-labs/indra/pkg/util/slice"
	"github.com/indra-labs/indra/pkg/util/splice"
	"net/netip"
	"reflect"
)

const (
	ReverseMagic = "rvrs"
	ReverseLen   = magic.Len + 1 + splice.AddrLen
)

type Reverse struct {
	AddrPort *netip.AddrPort
	Onion
}

func (x *Reverse) Account(res *sess.Data, sm *sess.Manager,
	s *sessions.Data, last bool) (skip bool, sd *sessions.Data) {
	res.Billable = append(res.Billable, s.Header.Bytes)
	return
}

func (x *Reverse) Decode(s *splice.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), ReverseLen-magic.Len,
		ReverseMagic); fails(e) {
		return
	}
	s.ReadAddrPort(&x.AddrPort)
	return
}

func (x *Reverse) Encode(s *splice.Splice) (e error) {
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

func (x *Reverse) GetOnion() interface{} { return x }

func (x *Reverse) Handle(s *splice.Splice, p Onion, ng Ngin) (e error) {
	if x.AddrPort.String() == ng.Mgr().GetLocalNodeAddress().String() {
		in := reg.Recognise(s)
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
		hdr, pld, _, _ := ng.Mgr().FindCloaked(on.Cloak)
		if hdr == nil || pld == nil {
			log.E.F("failed to find key for %s",
				ng.Mgr().GetLocalNodeAddressString())
			return e
		}
		// We need to find the PayloadPub to match.
		on.ToPriv = hdr
		// Decrypt using the Payload key and header nonce.
		c := s.GetCursor()
		ciph.Encipher(ciph.GetBlock(on.ToPriv, on.FromPub, "reverse header"),
			on.IV, s.GetRange(c, c+2*ReverseCryptLen))
		// shift the header segment upwards and pad the
		// remainder.
		s.CopyRanges(start, first, first, second)
		s.CopyRanges(first, second, second, last)
		s.CopyIntoRange(slice.NoisePad(ReverseCryptLen), second, last)
		if last != s.Len() {
			ciph.Encipher(ciph.GetBlock(pld, on.FromPub, "reverse payload"),
				on.IV, s.GetFrom(last))
		}
		if string(s.GetRange(start, start+magic.Len)) != ReverseMagic {
			// It's for us!
			log.T.S("handling response")
			ng.HandleMessage(splice.BudgeUp(s.SetCursor(last)), on)
			return e
		}
		sess := ng.Mgr().FindSessionByHeader(hdr)
		if sess != nil {
			ng.Mgr().DecSession(sess.Header.Bytes,
				ng.Mgr().GetLocalNodeRelayRate()*s.Len(), false, "reverse")
			ng.HandleMessage(splice.BudgeUp(s.SetCursor(start)), on)
		}
	} else if p != nil {
		// we need to forward this message onion.
		log.T.Ln(ng.Mgr().GetLocalNodeAddressString(), "forwarding reverse")
		ng.Mgr().Send(x.AddrPort, s)
	} else {
		log.E.Ln("we do not forward nonsense! scoff! snort!")
	}
	return e
}

func (x *Reverse) Len() int         { return ReverseLen + x.Onion.Len() }
func (x *Reverse) Magic() string    { return ReverseMagic }
func (x *Reverse) Wrap(inner Onion) { x.Onion = inner }
func init()                         { reg.Register(ReverseMagic, reverseGen) }
func reverseGen() coding.Codec      { return &Reverse{} }
