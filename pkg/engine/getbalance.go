package engine

import (
	"reflect"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/engine/magic"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

const (
	GetBalanceMagic = "gb"
	GetBalanceLen   = magic.Len + 2*nonce.IDLen +
		3*sha256.Len + nonce.IVLen*3
)

type GetBalance struct {
	ID     nonce.ID
	ConfID nonce.ID
	// Ciphers is a set of 3 symmetric ciphers that are to be used in their
	// given order over the reply message from the service.
	Ciphers
	// Nonces are the nonces to use with the cipher when creating the
	// encryption for the reply message,
	// they are common with the crypts in the header.
	Nonces
	// Port identifies the type of service as well as being the port used by
	// the service to be relayed to. Notice there is no IP address, this is
	// because Indranet only forwards to exits of decentralised services
	// also running on the same machine. This service could be a proxy, of
	// course, if configured this way. This could be done by tunneling from
	// a local Socks5 proxy into Indranet and the exit node also having
	// this.
	Onion
}

func getBalanceGen() Codec             { return &GetBalance{} }
func init()                            { Register(GetBalanceMagic, getBalanceGen) }
func (x *GetBalance) Magic() string    { return GetBalanceMagic }
func (x *GetBalance) Len() int         { return GetBalanceLen + x.Onion.Len() }
func (x *GetBalance) Wrap(inner Onion) { x.Onion = inner }
func (x *GetBalance) GetOnion() Onion  { return x }

type GetBalanceParams struct {
	ID, ConfID nonce.ID
	Alice, Bob *SessionData
	S          Circuit
	KS         *crypto.KeySet
}

// MakeGetBalance sends out a request in a similar way to Exit except the node
// being queried can be any of the 5.
func MakeGetBalance(p GetBalanceParams) Skins {
	headers := GetHeaders(p.Alice, p.Bob, p.S, p.KS)
	return Skins{}.
		RoutingHeader(headers.Forward).
		GetBalance(p.ID, p.ConfID, headers.ExitPoint()).
		RoutingHeader(headers.Return)
}

func (ng *Engine) SendGetBalance(alice, bob *SessionData, hook Callback) {
	hops := StandardCircuit()
	s := make(Sessions, len(hops))
	s[2] = bob
	s[5] = alice
	se := ng.SelectHops(hops, s, "sendgetbalance")
	var c Circuit
	copy(c[:], se)
	confID := nonce.NewID()
	o := MakeGetBalance(GetBalanceParams{alice.ID, confID, alice, bob, c,
		ng.KeySet})
	log.D.Ln("sending out getbalance onion")
	res := ng.PostAcctOnion(o)
	ng.SendWithOneHook(c[0].Node.AddrPort, res, hook, ng.PendingResponses)
}

func (o Skins) GetBalance(id, confID nonce.ID, ep *ExitPoint) Skins {
	return append(o, &GetBalance{
		ID:      id,
		ConfID:  confID,
		Ciphers: GenCiphers(ep.Keys, ep.ReturnPubs),
		Nonces:  ep.Nonces,
		Onion:   nop,
	})
}

func (x *GetBalance) Encode(s *Splice) (e error) {
	log.T.S("encoding", reflect.TypeOf(x),
		x.ID, x.ConfID, x.Ciphers, x.Nonces,
	)
	return x.Onion.Encode(s.
		Magic(GetBalanceMagic).
		ID(x.ID).
		ID(x.ConfID).Ciphers(x.Ciphers).Nonces(x.Nonces),
	)
}

func (x *GetBalance) Decode(s *Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), GetBalanceLen-magic.Len,
		GetBalanceMagic); fails(e) {
		return
	}
	s.ReadID(&x.ID).
		ReadID(&x.ConfID).ReadCiphers(&x.Ciphers).ReadNonces(&x.Nonces)
	return
}

func (x *GetBalance) Handle(s *Splice, p Onion,
	ng *Engine) (e error) {
	
	log.T.S(x)
	var found bool
	var bal *Balance
	ng.IterateSessions(func(sd *SessionData) bool {
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
		log.D.S(ng.Sessions)
		return
	}
	log.D.Ln("session found", x.ID)
	header := s.GetRoutingHeaderFromCursor()
	rbb := FormatReply(header, x.Ciphers, x.Nonces, Encode(bal).GetAll())
	rb := append(rbb.GetAll(), slice.NoisePad(714-rbb.Len())...)
	switch on1 := p.(type) {
	case *Crypt:
		sess := ng.FindSessionByHeader(on1.ToPriv)
		if sess != nil {
			in := sess.Node.RelayRate * s.Len() / 2
			out := sess.Node.RelayRate * len(rb) / 2
			ng.DecSession(sess.ID, in+out, false, "getbalance")
		}
	}
	ng.IterateSessions(func(sd *SessionData) bool {
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
	ng.HandleMessage(LoadSplice(rb, slice.NewCursor()), x)
	return
}

func (x *GetBalance) Account(res *SendData, sm *SessionManager,
	s *SessionData, last bool) (skip bool, sd *SessionData) {
	
	res.ID = s.ID
	res.Billable = append(res.Billable, s.ID)
	skip = true
	return
}
