package ready

import (
	"github.com/indra-labs/indra"
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/ciph"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/crypto/sha256"
	"github.com/indra-labs/indra/pkg/engine/coding"
	"github.com/indra-labs/indra/pkg/engine/magic"
	"github.com/indra-labs/indra/pkg/engine/sess"
	"github.com/indra-labs/indra/pkg/engine/sessions"
	"github.com/indra-labs/indra/pkg/onions/consts"
	"github.com/indra-labs/indra/pkg/onions/hidden"
	"github.com/indra-labs/indra/pkg/onions/ont"
	"github.com/indra-labs/indra/pkg/onions/reg"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/splice"
	"reflect"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	fails = log.E.Chk
)

const (
	ReadyMagic = "redy"
	ReadyLen   = magic.Len + nonce.IDLen + crypto.PubKeyLen + 2*consts.RoutingHeaderLen +
		3*sha256.Len + 3*nonce.IVLen
)

type Ready struct {
	ID              nonce.ID
	Address         *crypto.Pub
	Forward, Return *hidden.ReplyHeader
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
	x.Return = &hidden.ReplyHeader{}
	s.ReadID(&x.ID).
		ReadPubkey(&x.Address)
	hidden.ReadRoutingHeader(s, &x.Return.RoutingHeaderBytes).
		ReadCiphers(&x.Return.Ciphers).
		ReadNonces(&x.Return.Nonces)
	return
}

func (x *Ready) Encode(s *splice.Splice) (e error) {
	log.T.S("encoding", reflect.TypeOf(x),
		x.ID, x.Address, x.Forward,
	)
	hidden.WriteRoutingHeader(s, x.Forward.RoutingHeaderBytes)
	start := s.GetCursor()
	s.Magic(ReadyMagic).
		ID(x.ID).
		Pubkey(x.Address)
	hidden.WriteRoutingHeader(s, x.Return.RoutingHeaderBytes).
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

func (x *Ready) Handle(s *splice.Splice, p ont.Onion, ng ont.Ngin) (e error) {
	_, e = ng.Pending().ProcessAndDelete(x.ID, x, s.GetAll())
	return
}

func (x *Ready) Len() int             { return ReadyLen }
func (x *Ready) Magic() string        { return ReadyMagic }
func (x *Ready) Wrap(inner ont.Onion) {}

func NewReady(
	id nonce.ID,
	addr *crypto.Pub,
	fwHeader,
	rvHeader hidden.RoutingHeaderBytes,
	fc, rc crypto.Ciphers,
	fn, rn crypto.Nonces,
) ont.Onion {return &Ready{id, addr,
	&hidden.ReplyHeader{fwHeader, fc, fn},
	&hidden.ReplyHeader{rvHeader, rc, rn},
}}

func init()                       { reg.Register(ReadyMagic, readyGen) }
func readyGen() coding.Codec { return &Ready{} }
