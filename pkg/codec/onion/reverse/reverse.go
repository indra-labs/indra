// Package reverse provides a message type for the forwarding directions in a 3 layer routing header.
package reverse

import (
	"github.com/indra-labs/indra/pkg/codec"
	"github.com/indra-labs/indra/pkg/codec/onion/crypt"
	"github.com/indra-labs/indra/pkg/codec/onion/end"
	"github.com/indra-labs/indra/pkg/codec/ont"
	reg2 "github.com/indra-labs/indra/pkg/codec/reg"
	"github.com/indra-labs/indra/pkg/crypto/ciph"
	"github.com/indra-labs/indra/pkg/engine/consts"
	"github.com/indra-labs/indra/pkg/engine/magic"
	"github.com/indra-labs/indra/pkg/engine/sess"
	"github.com/indra-labs/indra/pkg/engine/sessions"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/slice"
	"github.com/indra-labs/indra/pkg/util/splice"
	"net/netip"
	"reflect"
)

var (
	log   = log2.GetLogger()
	fails = log.E.Chk
)

const (
	ReverseMagic = "rvrs"
)

type Reverse struct {
	AddrPort *netip.AddrPort
	ont.Onion
}

func NewReverse(ip *netip.AddrPort) ont.Onion {
	return &Reverse{AddrPort: ip, Onion: end.NewEnd()}
}

func (x *Reverse) Account(res *sess.Data, sm *sess.Manager,
	s *sessions.Data, last bool) (skip bool, sd *sessions.Data) {
	res.Billable = append(res.Billable, s.Header.Bytes)
	return
}

func (x *Reverse) Decode(s *splice.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), consts.ReverseLen-magic.Len,
		ReverseMagic); fails(e) {
		return
	}
	s.ReadAddrPort(&x.AddrPort)
	return
}

func (x *Reverse) Encode(s *splice.Splice) (e error) {
	log.T.Ln("encoding", reflect.TypeOf(x), x.AddrPort)
	if x.AddrPort == nil {
		s.Advance(consts.ReverseLen, "reverse")
	} else {
		s.Magic(ReverseMagic).AddrPort(x.AddrPort)
	}
	if x.Onion != nil {
		e = x.Onion.Encode(s)
	}
	return
}

func (x *Reverse) GetOnion() interface{} { return x }

func (x *Reverse) Handle(s *splice.Splice, p ont.Onion, ng ont.Ngin) (e error) {
	if x.AddrPort.String() == ng.Mgr().GetLocalNodeAddress().String() {
		in := reg2.Recognise(s)
		if e = in.Decode(s); fails(e) {
			return e
		}
		if in.Magic() != crypt.CryptMagic {
			return e
		}
		on := in.(*crypt.Crypt)
		first := s.GetCursor()
		start := first - consts.ReverseCryptLen
		second := first + consts.ReverseCryptLen
		last := second + consts.ReverseCryptLen
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
			on.IV, s.GetRange(c, c+2*consts.ReverseCryptLen))
		// shift the header segment upwards and pad the
		// remainder.
		s.CopyRanges(start, first, first, second)
		s.CopyRanges(first, second, second, last)
		s.CopyIntoRange(slice.NoisePad(consts.ReverseCryptLen), second, last)
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
				int(ng.Mgr().GetLocalNodeRelayRate())*s.Len(), false, "reverse")
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

func (x *Reverse) Len() int             { return consts.ReverseLen + x.Onion.Len() }
func (x *Reverse) Magic() string        { return ReverseMagic }
func (x *Reverse) Wrap(inner ont.Onion) { x.Onion = inner }
func init()                             { reg2.Register(ReverseMagic, reverseGen) }
func reverseGen() codec.Codec           { return &Reverse{} }
