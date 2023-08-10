// Package reverse provides a message type for the forwarding directions in a 3 layer routing header.
package reverse

import (
	"git.indra-labs.org/dev/ind/pkg/cfg"
	"git.indra-labs.org/dev/ind/pkg/codec"
	"git.indra-labs.org/dev/ind/pkg/codec/onion/cores/end"
	"git.indra-labs.org/dev/ind/pkg/codec/onion/crypt"
	"git.indra-labs.org/dev/ind/pkg/codec/ont"
	reg2 "git.indra-labs.org/dev/ind/pkg/codec/reg"
	"git.indra-labs.org/dev/ind/pkg/crypto/ciph"
	"git.indra-labs.org/dev/ind/pkg/engine/consts"
	"git.indra-labs.org/dev/ind/pkg/engine/magic"
	"git.indra-labs.org/dev/ind/pkg/engine/sess"
	"git.indra-labs.org/dev/ind/pkg/engine/sessions"
	log2 "git.indra-labs.org/dev/ind/pkg/proc/log"
	"git.indra-labs.org/dev/ind/pkg/util/slice"
	"git.indra-labs.org/dev/ind/pkg/util/splice"
	"github.com/multiformats/go-multiaddr"
	"reflect"
)

var (
	log   = log2.GetLogger()
	fails = log.E.Chk
)

const (
	Magic = "rvrs"
)

// Reverse is a part of the 3 layer relay RoutingHeader which provides the next
// address to forward to.
//
// Deprecated: In process of obsoleting this and replacing with offset
type Reverse struct {

	// AddrPort of the relay to forward this message.
	Multiaddr multiaddr.Multiaddr

	// Onion contained inside this message.
	ont.Onion
}

// New creates a new Reverse onion.
func New(ip multiaddr.Multiaddr) ont.Onion {
	return &Reverse{Multiaddr: ip, Onion: end.NewEnd()}
}

// Account for the reverse message - note that the actual size being carried is
// computed at the end of the circuit with the returned
// Response/Confirmation/Message.
func (x *Reverse) Account(res *sess.Data, sm *sess.Manager,
	s *sessions.Data, last bool) (skip bool, sd *sessions.Data) {
	res.Billable = append(res.Billable, s.Header.Bytes)
	return
}

// Decode a Reverse from a provided splice.Splice.
func (x *Reverse) Decode(s *splice.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), consts.ReverseLen-magic.Len,
		Magic); fails(e) {
		return
	}
	s.ReadMultiaddr(&x.Multiaddr)
	return
}

// Encode a Reverse into the next bytes of a splice.Splice.
func (x *Reverse) Encode(s *splice.Splice) (e error) {
	log.T.Ln("encoding", reflect.TypeOf(x), x.Multiaddr)
	if x.Multiaddr == nil {
		s.Advance(consts.ReverseLen, "reverse")
	} else {
		s.Magic(Magic).
			Multiaddr(x.Multiaddr, cfg.GetCurrentDefaultPort())
	}
	if x.Onion != nil {
		e = x.Onion.Encode(s)
	}
	return
}

// Unwrap returns the onion inside this Reverse.
func (x *Reverse) Unwrap() interface{} { return x.Onion }

// Handle is the relay logic for an engine handling a Reverse message.
//
// This is where the 3 layer RoutingHeader is processed.
func (x *Reverse) Handle(s *splice.Splice, p ont.Onion, ng ont.Ngin) (e error) {
	if ng.Mgr().MatchesLocalNodeAddress(x.Multiaddr) {
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
		if string(s.GetRange(start, start+magic.Len)) != Magic {
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
		ng.Mgr().Send(x.Multiaddr, s)
	} else {
		log.E.Ln("we do not forward nonsense! scoff! snort!")
	}
	return e
}

// Len returns the length of this Reverse message.
func (x *Reverse) Len() int {

	codec.MustNotBeNil(x)

	// b, _ := multi.AddrToBytes(x.Multiaddr,
	// 	cfg.GetCurrentDefaultPort())

	return magic.Len + 21 + x.Onion.Len()
}

// Magic is the identifying 4 byte string indicating a Reverse message follows.
func (x *Reverse) Magic() string { return Magic }

// Wrap puts another onion inside this Reverse onion.
func (x *Reverse) Wrap(inner ont.Onion) { x.Onion = inner }

func init() { reg2.Register(Magic, Gen) }

// Gen is a factory function for a Reverse.
func Gen() codec.Codec { return &Reverse{} }
