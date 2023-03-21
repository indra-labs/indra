package ngin

import (
	"git-indra.lan/indra-labs/indra/pkg/crypto/ciph"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/ngin/magic"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

const (
	ReadyMagic = "rd"
	ReadyLen   = magic.Len + nonce.
		IDLen + RoutingHeaderLen // + 2*ForwardLen + 2*CryptLen
)

func ReadyPrototype() Onion { return &Ready{} }

func init() { Register(ReadyMagic, ReadyPrototype) }

type Ready struct {
	Reply         *Reply
	RoutingHeader slice.Bytes
}

func (o Skins) Ready(header slice.Bytes, reply *Reply) Skins {
	return append(o, &Ready{reply, header})
}

func (x *Ready) Magic() string { return ReadyMagic }

func (x *Ready) Encode(s *Splice) (e error) {
	s.RoutingHeader(x.RoutingHeader)
	start := s.GetCursor()
	s.Magic(ReadyMagic).ID(x.Reply.ID)
	for i := range x.Reply.Ciphers {
		blk := ciph.BlockFromHash(x.Reply.Ciphers[i])
		ciph.Encipher(blk, x.Reply.Nonces[2-i], s.GetFrom(start))
	}
	return
}

func (x *Ready) Decode(s *Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), ReadyLen-magic.Len,
		ReadyMagic); check(e) {
		return
	}
	x.Reply = &Reply{}
	s.ReadID(&x.Reply.ID)
	return
}

func (x *Ready) Len() int { return ReadyLen }

func (x *Ready) Wrap(inner Onion) {}

func (x *Ready) Handle(s *Splice, p Onion,
	ng *Engine) (e error) {
	
	log.D.Ln(ng.GetLocalNodeAddressString(), x.Reply.ID)
	ng.PendingResponses.ProcessAndDelete(x.Reply.ID, nil, s.GetAll())
	return
}

func MakeReady(header slice.Bytes, reply *Reply) *Ready {
	return &Ready{reply, header}
}
