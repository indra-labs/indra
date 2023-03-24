package engine

import (
	"reflect"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/ciph"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/engine/magic"
)

const (
	ReadyMagic = "rd"
	ReadyLen   = magic.Len + nonce.IDLen + pub.KeyLen + 2*RoutingHeaderLen +
		3*sha256.Len + 3*nonce.IVLen
)

func ReadyPrototype() Onion       { return &Ready{} }
func init()                       { Register(ReadyMagic, ReadyPrototype) }
func (x *Ready) Magic() string    { return ReadyMagic }
func (x *Ready) Len() int         { return ReadyLen }
func (x *Ready) Wrap(inner Onion) {}

type Ready struct {
	ID              nonce.ID
	Address         *pub.Key
	Forward, Return *ReplyHeader
}

func (o Skins) Ready(id nonce.ID, addr *pub.Key, fwHeader,
	rvHeader RoutingHeaderBytes,
	fc, rc Ciphers, fn, rn Nonces) Skins {
	return append(o, &Ready{id, addr,
		&ReplyHeader{fwHeader, fc, fn},
		&ReplyHeader{rvHeader, rc, rn},
	})
}

func (x *Ready) Encode(s *Splice) (e error) {
	log.T.S("encoding", reflect.TypeOf(x),
		x.ID, x.Address, x.Forward,
	)
	s.
		RoutingHeader(x.Forward.RoutingHeaderBytes)
	start := s.GetCursor()
	s.
		Magic(ReadyMagic).
		ID(x.ID).
		Pubkey(x.Address).
		RoutingHeader(x.Return.RoutingHeaderBytes).
		Ciphers(x.Return.Ciphers).
		Nonces(x.Return.Nonces)
	for i := range x.Forward.Ciphers {
		blk := ciph.BlockFromHash(x.Forward.Ciphers[i])
		log.D.F("encrypting %s", x.Forward.Ciphers[i].String())
		ciph.Encipher(blk, x.Forward.Nonces[i], s.GetFrom(start))
	}
	return
}

func (x *Ready) Decode(s *Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), ReadyLen-magic.Len,
		ReadyMagic); check(e) {
		return
	}
	x.Return = &ReplyHeader{}
	s.
		ReadID(&x.ID).
		ReadPubkey(&x.Address).
		ReadRoutingHeader(&x.Return.RoutingHeaderBytes).
		ReadCiphers(&x.Return.Ciphers).
		ReadNonces(&x.Return.Nonces)
	return
}

func (x *Ready) Handle(s *Splice, p Onion,
	ng *Engine) (e error) {
	
	_, e = ng.PendingResponses.ProcessAndDelete(x.ID, x, s.GetAll())
	return
}
