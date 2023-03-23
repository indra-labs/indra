package engine

import (
	"git-indra.lan/indra-labs/indra/pkg/crypto/ciph"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/engine/magic"
	"git-indra.lan/indra-labs/indra/pkg/engine/types"
)

const (
	ReadyMagic = "rd"
	ReadyLen   = magic.Len + nonce.IDLen + 2*RoutingHeaderLen +
		3*sha256.Len + 3*nonce.IVLen
)

func ReadyPrototype() Onion { return &Ready{} }

func init() { Register(ReadyMagic, ReadyPrototype) }

type ReplyHeader struct {
	RoutingHeaderBytes
	// Ciphers is a set of 3 symmetric ciphers that are to be used in their
	// given order over the reply message from the service.
	types.Ciphers
	// Nonces are the nonces to use with the cipher when creating the
	// encryption for the reply message,
	// they are common with the crypts in the header.
	types.Nonces
}

type Ready struct {
	ID               nonce.ID
	Forward, Reverse ReplyHeader
}

func (o Skins) Ready(id nonce.ID, fwHeader, rvHeader RoutingHeaderBytes,
	fc, rc types.Ciphers, fn, rn types.Nonces) Skins {
	return append(o, &Ready{id,
		ReplyHeader{fwHeader, fc, fn},
		ReplyHeader{rvHeader, rc, rn},
	})
}

func (x *Ready) Magic() string { return ReadyMagic }

func (x *Ready) Encode(s *Splice) (e error) {
	s.RoutingHeader(x.Forward.RoutingHeaderBytes)
	start := s.GetCursor()
	s.Magic(ReadyMagic).
		ID(x.ID).
		RoutingHeader(x.Reverse.RoutingHeaderBytes).
		Ciphers(x.Reverse.Ciphers).
		Nonces(x.Reverse.Nonces)
	for i := range x.Forward.Ciphers {
		blk := ciph.BlockFromHash(x.Forward.Ciphers[i])
		ciph.Encipher(blk, x.Forward.Nonces[2-i], s.GetFrom(start))
	}
	return
}

func (x *Ready) Decode(s *Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), ReadyLen-magic.Len,
		ReadyMagic); check(e) {
		return
	}
	s.
		ReadID(&x.ID).
		ReadRoutingHeader(&x.Reverse.RoutingHeaderBytes).
		ReadCiphers(&x.Reverse.Ciphers).
		ReadNonces(&x.Reverse.Nonces)
	return
}

func (x *Ready) Len() int { return ReadyLen }

func (x *Ready) Wrap(inner Onion) {}

func (x *Ready) Handle(s *Splice, p Onion,
	ng *Engine) (e error) {
	
	log.D.Ln(ng.GetLocalNodeAddressString(), x.ID)
	log.T.S("ready", x.Reverse.RoutingHeaderBytes, x.Reverse)
	ng.PendingResponses.ProcessAndDelete(x.ID, nil, s.GetAll())
	return
}
