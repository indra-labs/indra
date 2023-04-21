package onions

import (
	"reflect"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/engine/coding"
	"git-indra.lan/indra-labs/indra/pkg/engine/magic"
	"git-indra.lan/indra-labs/indra/pkg/engine/sess"
	"git-indra.lan/indra-labs/indra/pkg/engine/sessions"
	"git-indra.lan/indra-labs/indra/pkg/splice"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

const (
	GetBalanceMagic = "getb"
	GetBalanceLen   = magic.Len + 2*nonce.IDLen +
		3*sha256.Len + nonce.IVLen*3
)

type GetBalance struct {
	ID     nonce.ID
	ConfID nonce.ID
	// Ciphers is a set of 3 symmetric ciphers that are to be used in their
	// given order over the reply message from the service.
	crypto.Ciphers
	// Nonces are the nonces to use with the cipher when creating the
	// encryption for the reply message,
	// they are common with the crypts in the header.
	crypto.Nonces
	// Port identifies the type of service as well as being the port used by
	// the service to be relayed to. Notice there is no IP address, this is
	// because Indranet only forwards to exits of decentralised services
	// also running on the same machine. This service could be a proxy, of
	// course, if configured this way. This could be done by tunneling from
	// a local Socks5 proxy into Indranet and the exit node also having
	// this.
	Onion
}

func getBalanceGen() coding.Codec           { return &GetBalance{} }
func init()                                 { Register(GetBalanceMagic, getBalanceGen) }
func (x *GetBalance) Magic() string         { return GetBalanceMagic }
func (x *GetBalance) Len() int              { return GetBalanceLen + x.Onion.Len() }
func (x *GetBalance) Wrap(inner Onion)      { x.Onion = inner }
func (x *GetBalance) GetOnion() interface{} { return x }

type GetBalanceParams struct {
	ID, ConfID nonce.ID
	Alice, Bob *sessions.Data
	S          sessions.Circuit
	KS         *crypto.KeySet
}

func (x *GetBalance) Encode(s *splice.Splice) (e error) {
	log.T.S("encoding", reflect.TypeOf(x),
		x.ID, x.ConfID, x.Ciphers, x.Nonces,
	)
	return x.Onion.Encode(s.
		Magic(GetBalanceMagic).
		ID(x.ID).
		ID(x.ConfID).Ciphers(x.Ciphers).Nonces(x.Nonces),
	)
}

func (x *GetBalance) Decode(s *splice.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), GetBalanceLen-magic.Len,
		GetBalanceMagic); fails(e) {
		return
	}
	s.ReadID(&x.ID).
		ReadID(&x.ConfID).ReadCiphers(&x.Ciphers).ReadNonces(&x.Nonces)
	return
}

func (x *GetBalance) Handle(s *splice.Splice, p Onion, ng Ngin) (e error) {
	log.T.S(x)
	var found bool
	var bal *Balance
	ng.Mgr().IterateSessions(func(sd *sessions.Data) bool {
		if sd.ID == x.ID {
			log.D.S("sessiondata", sd.ID, sd.Remaining)
			bal = &Balance{
				ID:           x.ID,
				ConfID:       x.ConfID,
				MilliSatoshi: sd.Remaining,
			}
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
	header := GetRoutingHeaderFromCursor(s)
	rbb := FormatReply(header, x.Ciphers, x.Nonces, Encode(bal).GetAll())
	rb := append(rbb.GetAll(), slice.NoisePad(714-rbb.Len())...)
	switch on1 := p.(type) {
	case *Crypt:
		sess := ng.Mgr().FindSessionByHeader(on1.ToPriv)
		if sess != nil {
			in := sess.Node.RelayRate * s.Len() / 2
			out := sess.Node.RelayRate * len(rb) / 2
			ng.Mgr().DecSession(sess.ID, in+out, false, "getbalance")
		}
	}
	ng.Mgr().IterateSessions(func(sd *sessions.Data) bool {
		if sd.ID == x.ID {
			bal = &Balance{
				ID:           x.ID,
				ConfID:       x.ConfID,
				MilliSatoshi: sd.Remaining,
			}
			found = true
			return true
		}
		return false
	})
	rbb = FormatReply(header, x.Ciphers, x.Nonces, Encode(bal).GetAll())
	rb = append(rbb.GetAll(), slice.NoisePad(714-len(rb))...)
	ng.HandleMessage(splice.Load(rb, slice.NewCursor()), x)
	return
}

func (x *GetBalance) Account(res *sess.Data, sm *sess.Manager,
	s *sessions.Data, last bool) (skip bool, sd *sessions.Data) {
	
	res.ID = s.ID
	res.Billable = append(res.Billable, s.ID)
	skip = true
	return
}
