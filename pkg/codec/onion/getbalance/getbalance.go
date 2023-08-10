// Package getbalance provides an onion message layer type that makes a request for the current balance of a session.
package getbalance

import (
	"git.indra-labs.org/dev/ind/pkg/codec"
	"git.indra-labs.org/dev/ind/pkg/codec/onion/cores/balance"
	"git.indra-labs.org/dev/ind/pkg/codec/onion/cores/end"
	"git.indra-labs.org/dev/ind/pkg/codec/onion/crypt"
	"git.indra-labs.org/dev/ind/pkg/codec/onion/exit"
	"git.indra-labs.org/dev/ind/pkg/codec/ont"
	"git.indra-labs.org/dev/ind/pkg/codec/reg"
	"git.indra-labs.org/dev/ind/pkg/crypto"
	"git.indra-labs.org/dev/ind/pkg/crypto/nonce"
	"git.indra-labs.org/dev/ind/pkg/crypto/sha256"
	"git.indra-labs.org/dev/ind/pkg/engine/magic"
	"git.indra-labs.org/dev/ind/pkg/engine/sess"
	"git.indra-labs.org/dev/ind/pkg/engine/sessions"
	"git.indra-labs.org/dev/ind/pkg/hidden"
	log2 "git.indra-labs.org/dev/ind/pkg/proc/log"
	"git.indra-labs.org/dev/ind/pkg/util/slice"
	"git.indra-labs.org/dev/ind/pkg/util/splice"
	"reflect"
)

var (
	log   = log2.GetLogger()
	fails = log.E.Chk
)

const (
	GetBalanceMagic = "getb"
	GetBalanceLen   = magic.Len + nonce.IDLen +
		3*sha256.Len + nonce.IVLen*3
)

type GetBalance struct {
	ID nonce.ID

	// Ciphers is a set of 3 symmetric ciphers that are to be used in their
	// given order over the reply message from the service.
	crypto.Ciphers

	// Nonces are the nonces to use with the cipher when creating the
	// encryption for the reply message,
	// they are common with the crypts in the header.
	crypto.Nonces

	ont.Onion
}

func (x *GetBalance) Account(res *sess.Data, sm *sess.Manager,
	s *sessions.Data, last bool) (skip bool, sd *sessions.Data) {
	res.ID = x.ID
	res.Billable = append(res.Billable, s.Header.Bytes)
	skip = true
	return
}

func (x *GetBalance) Decode(s *splice.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), GetBalanceLen-magic.Len,
		GetBalanceMagic); fails(e) {
		return
	}
	s.ReadID(&x.ID).
		ReadCiphers(&x.Ciphers).ReadNonces(&x.Nonces)
	return
}

func (x *GetBalance) Encode(s *splice.Splice) (e error) {
	log.T.S("encoding", reflect.TypeOf(x),
		x.ID, x.Ciphers, x.Nonces,
	)
	return x.Onion.Encode(s.
		Magic(GetBalanceMagic).
		ID(x.ID).
		Ciphers(x.Ciphers).Nonces(x.Nonces),
	)
}

func (x *GetBalance) Unwrap() interface{} { return x }

func (x *GetBalance) Handle(s *splice.Splice, p ont.Onion, ng ont.Ngin) (e error) {
	log.T.S(x)
	var found bool
	var bal *balance.Balance
	ng.Mgr().IterateSessions(func(sd *sessions.Data) bool {
		if sd.ID == x.ID {
			log.D.S("sessiondata", sd.ID, sd.Remaining)
			bal = &balance.Balance{ID: x.ID, MilliSatoshi: sd.Remaining}
			log.D.S("bal", bal)
			found = true
			return true
		}
		return false
	})
	if !found {
		log.E.Ln("session not found", x.ID)
		log.D.S(ng.Mgr().Sessions)
		return
	}
	log.D.Ln("session found", x.ID)
	header := hidden.GetRoutingHeaderFromCursor(s)
	rbb := hidden.FormatReply(header, x.Ciphers, x.Nonces, codec.Encode(bal).GetAll())
	rb := append(rbb.GetAll(), slice.NoisePad(714-rbb.Len())...)
	switch on1 := p.(type) {
	case *crypt.Crypt:
		sess := ng.Mgr().FindSessionByHeader(on1.ToPriv)
		if sess != nil {
			in := int(sess.Node.RelayRate) * s.Len() / 2
			out := int(sess.Node.RelayRate) * len(rb) / 2
			ng.Mgr().DecSession(sess.Header.Bytes, in+out, false, "getbalance")
		}
	}
	ng.Mgr().IterateSessions(func(sd *sessions.Data) bool {
		if sd.ID == x.ID {
			bal = &balance.Balance{ID: x.ID, MilliSatoshi: sd.Remaining}
			found = true
			return true
		}
		return false
	})
	rbb = hidden.FormatReply(header, x.Ciphers, x.Nonces, codec.Encode(bal).GetAll())
	rb = append(rbb.GetAll(), slice.NoisePad(714-len(rb))...)
	ng.HandleMessage(splice.Load(rb, slice.NewCursor()), x)
	return
}

func (x *GetBalance) Len() int {

	codec.MustNotBeNil(x)

	return GetBalanceLen + x.Onion.Len()
}

func (x *GetBalance) Magic() string { return GetBalanceMagic }

type GetBalanceParams struct {
	ID         nonce.ID
	Alice, Bob *sessions.Data
	S          sessions.Circuit
	KS         *crypto.KeySet
}

func (x *GetBalance) Wrap(inner ont.Onion) { x.Onion = inner }
func NewGetBalance(id nonce.ID, ep *exit.ExitPoint) ont.Onion {
	return &GetBalance{
		ID:      id,
		Ciphers: crypto.GenCiphers(ep.Keys, ep.ReturnPubs),
		Nonces:  ep.Nonces,
		Onion:   end.NewEnd(),
	}
}
func getBalanceGen() codec.Codec { return &GetBalance{} }
func init()                      { reg.Register(GetBalanceMagic, getBalanceGen) }
