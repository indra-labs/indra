package onions

import (
	"reflect"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto"
	"git-indra.lan/indra-labs/indra/pkg/crypto/ciph"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/engine/coding"
	"git-indra.lan/indra-labs/indra/pkg/engine/magic"
	"git-indra.lan/indra-labs/indra/pkg/engine/sess"
	"git-indra.lan/indra-labs/indra/pkg/engine/sessions"
	"git-indra.lan/indra-labs/indra/pkg/util/splice"
)

const (
	ReadyMagic = "redy"
	ReadyLen   = magic.Len + nonce.IDLen + crypto.PubKeyLen + 2*RoutingHeaderLen +
		3*sha256.Len + 3*nonce.IVLen
)

func ReadyGen() coding.Codec           { return &Ready{} }
func init()                            { Register(ReadyMagic, ReadyGen) }
func (x *Ready) Magic() string         { return ReadyMagic }
func (x *Ready) Len() int              { return ReadyLen }
func (x *Ready) Wrap(inner Onion)      {}
func (x *Ready) GetOnion() interface{} { return x }

type Ready struct {
	ID              nonce.ID
	Address         *crypto.Pub
	Forward, Return *ReplyHeader
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
		log.D.F("encrypting %s", x.Forward.Ciphers[i].Based32String())
		ciph.Encipher(blk, x.Forward.Nonces[i], s.GetFrom(start))
	}
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

func (x *Ready) Handle(s *splice.Splice, p Onion, ng Ngin) (e error) {
	// todo: this should be triggering connection open signal to socks/tunnel.
	_, e = ng.Pending().ProcessAndDelete(x.ID, x, s.GetAll())
	return
}

func (x *Ready) Account(res *sess.Data, sm *sess.Manager,
	s *sessions.Data, last bool) (skip bool, sd *sessions.Data) {
	
	return
}
