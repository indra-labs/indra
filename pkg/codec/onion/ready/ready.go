// Package ready provides an onion message type that is sent via client provided routing header back to the client after an introducer forwards a route message to initiate a hidden service connection.
package ready

import (
	"github.com/indra-labs/indra/pkg/codec"
	"github.com/indra-labs/indra/pkg/codec/ont"
	"github.com/indra-labs/indra/pkg/codec/reg"
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/ciph"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/crypto/sha256"
	"github.com/indra-labs/indra/pkg/engine/consts"
	"github.com/indra-labs/indra/pkg/engine/magic"
	"github.com/indra-labs/indra/pkg/engine/sess"
	"github.com/indra-labs/indra/pkg/engine/sessions"
	"github.com/indra-labs/indra/pkg/hidden"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/util/splice"
	"reflect"
)

var (
	log   = log2.GetLogger()
	fails = log.E.Chk
)

const (
	Magic = "redy"
	Len   = magic.Len + nonce.IDLen + crypto.PubKeyLen + 2*consts.RoutingHeaderLen +
		3*sha256.Len + 3*nonce.IVLen
)

// Ready is the connection confirmation after a client sends a Route to be referred forward to a hidden service.
//
// After this message the client can send a request to the hidden service.
type Ready struct {
	ID              nonce.ID
	Address         *crypto.Pub
	Forward, Return *hidden.ReplyHeader
}

// Account for a Ready onion - which is nothing. But it should signal to the
// tunnel/socket ready to receive.
func (x *Ready) Account(res *sess.Data, sm *sess.Manager,
	s *sessions.Data, last bool) (skip bool, sd *sessions.Data) {
	return
}

// Decode what should be an Ready message from a splice.Splice.
func (x *Ready) Decode(s *splice.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), Len-magic.Len,
		Magic); fails(e) {
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

// Encode this Ready onion into a splice.Splice's next bytes.
func (x *Ready) Encode(s *splice.Splice) (e error) {
	log.T.S("encoding", reflect.TypeOf(x),
		x.ID, x.Address, x.Forward,
	)
	hidden.WriteRoutingHeader(s, x.Forward.RoutingHeaderBytes)
	start := s.GetCursor()
	s.Magic(Magic).
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

// GetOnion is a no-op because there is no onion inside a Ready message.
func (x *Ready) GetOnion() interface{} { return nil }

// Handle provides the relay switching logic for an engine handling an Ready
// message.
func (x *Ready) Handle(s *splice.Splice, p ont.Onion, ng ont.Ngin) (e error) {
	_, e = ng.Pending().ProcessAndDelete(x.ID, x, s.GetAll())

	// todo: this should signal ready to send.
	return
}

// Len returns the length of the onion starting from this one.
func (x *Ready) Len() int { return Len }

// Magic bytes identifying a HiddenService message is up next.
func (x *Ready) Magic() string { return Magic }

// Wrap places another onion inside this one in its slot.
func (x *Ready) Wrap(inner ont.Onion) {}

// New generates a new Ready data structure and returns it as an ont.Onion
// interface.
func New(
	id nonce.ID,
	addr *crypto.Pub,
	fwHeader,
	rvHeader hidden.RoutingHeaderBytes,
	fc, rc crypto.Ciphers,
	fn, rn crypto.Nonces,
) ont.Onion {
	return &Ready{id, addr,
		&hidden.ReplyHeader{fwHeader, fc, fn},
		&hidden.ReplyHeader{rvHeader, rc, rn},
	}
}

// Gen is a factory function for an Ready.
func Gen() codec.Codec { return &Ready{} }

func init() { reg.Register(Magic, Gen) }
