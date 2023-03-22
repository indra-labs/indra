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

type Ready struct {
	FwReply  *Reply
	FwHeader slice.Bytes
	RvReply  *Reply
	RvHeader slice.Bytes
}

func (o Skins) Ready(fwHeader, rvHeader slice.Bytes, fwReply,
	rvReply *Reply) Skins {
	return append(o, &Ready{fwReply, fwHeader, rvReply, rvHeader})
}

func (x *Ready) Magic() string { return ReadyMagic }

func (x *Ready) Encode(s *Splice) (e error) {
	s.RoutingHeader(x.FwHeader)
	start := s.GetCursor()
	s.Magic(ReadyMagic).
		ID(x.FwReply.ID).
		Reply(x.RvReply).
		RoutingHeader(x.RvHeader)
	for i := range x.FwReply.Ciphers {
		blk := ciph.BlockFromHash(x.FwReply.Ciphers[i])
		ciph.Encipher(blk, x.FwReply.Nonces[2-i], s.GetFrom(start))
	}
	return
}

func (x *Ready) Decode(s *Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), ReadyLen-magic.Len,
		ReadyMagic); check(e) {
		return
	}
	x.FwReply = &Reply{}
	x.RvReply = &Reply{}
	s.ReadID(&x.FwReply.ID).
		ReadReply(x.RvReply).
		ReadRoutingHeader(&x.RvHeader)
	return
}

func (x *Ready) Len() int { return ReadyLen }

func (x *Ready) Wrap(inner Onion) {}

func (x *Ready) Handle(s *Splice, p Onion,
	ng *Engine) (e error) {
	
	log.D.Ln(ng.GetLocalNodeAddressString(), x.FwReply.ID)
	log.T.S("ready", x.RvHeader, x.RvReply)
	ng.PendingResponses.ProcessAndDelete(x.FwReply.ID, nil, s.GetAll())
	return
}
