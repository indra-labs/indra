package engine

import (
	"git-indra.lan/indra-labs/indra/pkg/crypto/ciph"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/engine/magic"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

const (
	MessageMagic = "ms"
	MessageLen   = 0
)

func MessagePrototype() Onion { return &Message{} }

func init() { Register(MessageMagic, MessagePrototype) }

type RouteCiphers struct {
	Header  slice.Bytes
	Ciphers [3]sha256.Hash
	IVs     [3]nonce.IV
}

type Message struct {
	ID              nonce.ID
	Forward, Return RouteCiphers
	Payload         slice.Bytes
}

func (o Skins) Message() Skins {
	return append(o, &Message{})
}

func NewMessage() *Message {
	return &Message{}
}

func (x *Message) Magic() string { return MessageMagic }

func (x *Message) Encode(s *Splice) (e error) {
	s.RoutingHeader(x.Forward.Header)
	start := s.GetCursor()
	s.Magic(MessageMagic).
		ID(x.ID).
		RoutingHeader(x.Return.Header).
		HashTriple(x.Return.Ciphers).
		IVTriple(x.Return.IVs).
		Bytes(x.Payload)
	for i := range x.Forward.Ciphers {
		blk := ciph.BlockFromHash(x.Forward.Ciphers[i])
		ciph.Encipher(blk, x.Forward.IVs[2-i], s.GetFrom(start))
	}
	return
}

func (x *Message) Decode(s *Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), ReadyLen-magic.Len,
		ReadyMagic); check(e) {
		return
	}
	s.ReadID(&x.ID).
		ReadRoutingHeader(&x.Return.Header).
		ReadHashTriple(&x.Return.Ciphers).
		ReadIVTriple(&x.Return.IVs)
	return
}

func (x *Message) Len() int { return MessageLen }

func (x *Message) Wrap(inner Onion) {}

func (x *Message) Handle(s *Splice, p Onion,
	ng *Engine) (e error) {
	
	return
}
