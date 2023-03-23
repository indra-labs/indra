package engine

import (
	"git-indra.lan/indra-labs/indra/pkg/crypto/ciph"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/engine/magic"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

const (
	ReadyMagic = "rd"
	ReadyLen   = magic.Len + nonce.
		IDLen + 2*RoutingHeaderLen + ReplyLen
)

func ReadyPrototype() Onion { return &Ready{} }

func init() { Register(ReadyMagic, ReadyPrototype) }

type ReplyHeader struct {
	Header slice.Bytes
	Reply  *Reply
}

type Ready struct {
	Forward, Reverse ReplyHeader
}

func (o Skins) Ready(fwHeader, rvHeader slice.Bytes, fwReply,
	rvReply *Reply) Skins {
	return append(o, &Ready{
		ReplyHeader{fwHeader, fwReply},
		ReplyHeader{rvHeader, rvReply},
	})
}

func (x *Ready) Magic() string { return ReadyMagic }

func (x *Ready) Encode(s *Splice) (e error) {
	s.RoutingHeader(x.Forward.Header)
	start := s.GetCursor()
	s.Magic(ReadyMagic).
		ID(x.Forward.Reply.ID).
		Ciphers(x.Reverse.Reply.Ciphers).
		Nonces(x.Reverse.Reply.Nonces).
		RoutingHeader(x.Reverse.Header)
	for i := range x.Forward.Reply.Ciphers {
		blk := ciph.BlockFromHash(x.Forward.Reply.Ciphers[i])
		ciph.Encipher(blk, x.Forward.Reply.Nonces[2-i], s.GetFrom(start))
	}
	return
}

func (x *Ready) Decode(s *Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), ReadyLen-magic.Len,
		ReadyMagic); check(e) {
		return
	}
	x.Forward.Reply = &Reply{}
	x.Reverse.Reply = &Reply{}
	s.
		ReadID(&x.Forward.Reply.ID).
		ReadCiphers(&x.Reverse.Reply.Ciphers).
		ReadNonces(&x.Reverse.Reply.Nonces).
		ReadRoutingHeader(&x.Reverse.Header)
	return
}

func (x *Ready) Len() int { return ReadyLen }

func (x *Ready) Wrap(inner Onion) {}

func (x *Ready) Handle(s *Splice, p Onion,
	ng *Engine) (e error) {
	
	log.D.Ln(ng.GetLocalNodeAddressString(), x.Forward.Reply.ID)
	log.T.S("ready", x.Reverse.Header, x.Reverse.Reply)
	ng.PendingResponses.ProcessAndDelete(x.Forward.Reply.ID, nil, s.GetAll())
	return
}
