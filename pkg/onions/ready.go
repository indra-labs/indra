package onions

import (
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/ciph"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/crypto/sha256"
	"github.com/indra-labs/indra/pkg/engine/coding"
	"github.com/indra-labs/indra/pkg/engine/magic"
	"github.com/indra-labs/indra/pkg/engine/sess"
	"github.com/indra-labs/indra/pkg/engine/sessions"
	"github.com/indra-labs/indra/pkg/onions/reg"
	"github.com/indra-labs/indra/pkg/util/splice"
	"reflect"
)

const (
	ReadyMagic = "redy"
	ReadyLen   = magic.Len + nonce.IDLen + crypto.PubKeyLen + 2*RoutingHeaderLen +
		3*sha256.Len + 3*nonce.IVLen
)

type Ready struct {
	ID              nonce.ID
	Address         *crypto.Pub
	Forward, Return *ReplyHeader
}

func (x *Ready) Account(res *sess.Data, sm *sess.Manager,
	s *sessions.Data, last bool) (skip bool, sd *sessions.Data) {
	return
}

func (x *Ready) Decode(s *splice.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), ReadyLen-magic.Len,
		ReadyMagic); fails(e) {
		return
	}
	x.Return = &ReplyHeader{}
	s.ReadID(&x.ID).
		ReadPubkey(&x.Address)
	ReadRoutingHeader(s, &x.Return.RoutingHeaderBytes).
		ReadCiphers(&x.Return.Ciphers).
		ReadNonces(&x.Return.Nonces)
	return
}

func (x *Ready) Encode(s *splice.Splice) (e error) {
	log.T.S("encoding", reflect.TypeOf(x),
		x.ID, x.Address, x.Forward,
	)
	WriteRoutingHeader(s, x.Forward.RoutingHeaderBytes)
	start := s.GetCursor()
	s.Magic(ReadyMagic).
		ID(x.ID).
		Pubkey(x.Address)
	WriteRoutingHeader(s, x.Return.RoutingHeaderBytes).
		Ciphers(x.Return.Ciphers).
		Nonces(x.Return.Nonces)
	for i := range x.Forward.Ciphers {
		blk := ciph.BlockFromHash(x.Forward.Ciphers[i])
		log.D.F("encrypting %s", x.Forward.Ciphers[i])
		ciph.Encipher(blk, x.Forward.Nonces[i], s.GetFrom(start))
	}
	return
}

func (x *Ready) GetOnion() interface{} { return x }

func (x *Ready) Handle(s *splice.Splice, p Onion, ng Ngin) (e error) {
	_, e = ng.Pending().ProcessAndDelete(x.ID, x, s.GetAll())
	return
}

func (x *Ready) Len() int         { return ReadyLen }
func (x *Ready) Magic() string    { return ReadyMagic }
func (x *Ready) Wrap(inner Onion) {}

func NewReady(
	id nonce.ID,
	addr *crypto.Pub,
	fwHeader,
	rvHeader RoutingHeaderBytes,
	fc, rc crypto.Ciphers,
	fn, rn crypto.Nonces,
)Onion{return &Ready{id, addr,
	&ReplyHeader{fwHeader, fc, fn},
	&ReplyHeader{rvHeader, rc, rn},
}}

func init()                       { reg.Register(ReadyMagic, readyGen) }
func readyGen() coding.Codec { return &Ready{} }
