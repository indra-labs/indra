package engine

import (
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/signer"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/engine/magic"
	"git-indra.lan/indra-labs/indra/pkg/util/octet"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

const (
	GetBalanceMagic = "gb"
	GetBalanceLen   = magic.Len + 2*nonce.IDLen +
		3*sha256.Len + nonce.IVLen*3
)

type GetBalance struct {
	nonce.ID
	octet.Reply
	Onion
}

func getBalancePrototype() Onion { return &GetBalance{} }

func init() { Register(GetBalanceMagic, getBalancePrototype) }

type GetBalanceParams struct {
	ID, ConfID nonce.ID
	Client     *SessionData
	S          Circuit
	KS         *signer.KeySet
}

// MakeGetBalance sends out a request in a similar way to Exit except the node
// being queried can be any of the 5.
func MakeGetBalance(p GetBalanceParams) Skins {
	headers := GetHeaders(p.Client, p.S, p.KS)
	return Skins{}.
		RoutingHeader(headers.Forward).
		GetBalance(p.ID, p.ConfID, headers.ExitPoint()).
		RoutingHeader(headers.Return)
}

func (ng *Engine) SendGetBalance(target *SessionData, hook Callback) {
	hops := StandardCircuit()
	s := make(Sessions, len(hops))
	s[2] = target
	se := ng.SelectHops(hops, s)
	var c Circuit
	copy(c[:], se)
	confID := nonce.NewID()
	o := MakeGetBalance(GetBalanceParams{target.ID, confID, se[5], c,
		ng.KeySet})
	log.D.Ln("sending out getbalance onion")
	res := ng.PostAcctOnion(o)
	ng.SendWithOneHook(c[0].AddrPort, res, hook, ng.PendingResponses)
}

func (o Skins) GetBalance(id, confID nonce.ID, ep *ExitPoint) Skins {
	
	return append(o, &GetBalance{
		ID: id,
		Reply: octet.Reply{
			ID:      confID,
			Ciphers: GenCiphers(ep.Keys, ep.ReturnPubs),
			Nonces:  ep.Nonces,
		},
		Onion: nop,
	})
}

func (x *GetBalance) Magic() string { return GetBalanceMagic }

func (x *GetBalance) Encode(s *octet.Splice) (e error) {
	return x.Onion.Encode(s.
		Magic(GetBalanceMagic).
		ID(x.ID).ID(x.Reply.ID).
		HashTriple(x.Ciphers).
		IVTriple(x.Nonces),
	)
}

func (x *GetBalance) Decode(s *octet.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), GetBalanceLen-magic.Len,
		GetBalanceMagic); check(e) {
		return
	}
	s.
		ReadID(&x.ID).ReadID(&x.Reply.ID).
		ReadHashTriple(&x.Ciphers).
		ReadIVTriple(&x.Nonces)
	return
}

func (x *GetBalance) Len() int { return GetBalanceLen + x.Onion.Len() }

func (x *GetBalance) Wrap(inner Onion) { x.Onion = inner }

func (x *GetBalance) Handle(s *octet.Splice, p Onion,
	ng *Engine) (e error) {
	
	log.T.S(x)
	var found bool
	var bal *Balance
	ng.IterateSessions(func(sd *SessionData) bool {
		if sd.ID == x.ID {
			log.D.S("sessiondata", sd.ID, sd.Remaining)
			bal = &Balance{
				ID:           x.ID,
				ConfID:       x.Reply.ID,
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
	header := s.GetRange(s.GetCursor(), s.Advance(ReverseHeaderLen))
	rbb := FormatReply(header,
		Encode(bal).GetRange(-1, -1), x.Ciphers, x.Nonces)
	rb := append(rbb.GetRange(-1, -1), slice.NoisePad(714-rbb.Len())...)
	switch on1 := p.(type) {
	case *Crypt:
		sess := ng.FindSessionByHeader(on1.ToPriv)
		if sess != nil {
			in := sess.RelayRate * s.Len() / 2
			out := sess.RelayRate * len(rb) / 2
			ng.DecSession(sess.ID, in+out, false, "getbalance")
		}
	}
	ng.IterateSessions(func(sd *SessionData) bool {
		if sd.ID == x.ID {
			bal = &Balance{
				ID:           x.ID,
				ConfID:       x.Reply.ID,
				MilliSatoshi: sd.Remaining,
			}
			found = true
			return true
		}
		return false
	})
	rbb = FormatReply(header,
		Encode(bal).GetRange(-1, -1), x.Ciphers, x.Nonces)
	rb = append(rbb.GetRange(-1, -1), slice.NoisePad(714-len(rb))...)
	ng.HandleMessage(octet.Load(rb, slice.NewCursor()), x)
	return
}
