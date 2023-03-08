package engine

import (
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
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
	ConfID nonce.ID
	// Ciphers is a set of 3 symmetric ciphers that are to be used in their
	// given order over the reply message from the service.
	Ciphers [3]sha256.Hash
	// Nonces are the nonces to use with the cipher when creating the
	// encryption for the reply message.
	Nonces [3]nonce.IV
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

// GetBalanceOnion sends out a request in a similar way to Exit except the node
// being queried can be any of the 5.
func GetBalanceOnion(p GetBalanceParams) Skins {
	var prvs [3]*prv.Key
	for i := range prvs {
		prvs[i] = p.KS.Next()
	}
	n := GenNonces(6)
	var retNonces [3]nonce.IV
	copy(retNonces[:], n[3:])
	var pubs [3]*pub.Key
	pubs[0] = p.S[3].PayloadPub
	pubs[1] = p.S[4].PayloadPub
	pubs[2] = p.Client.PayloadPub
	return Skins{}.
		ReverseCrypt(p.S[0], p.KS.Next(), n[0], 3).
		ReverseCrypt(p.S[1], p.KS.Next(), n[1], 2).
		ReverseCrypt(p.S[2], p.KS.Next(), n[2], 1).
		GetBalance(p.ID, p.ConfID, prvs, pubs, retNonces).
		ReverseCrypt(p.S[3], prvs[0], n[3], 0).
		ReverseCrypt(p.S[4], prvs[1], n[4], 0).
		ReverseCrypt(p.Client, prvs[2], n[5], 0)
}

func (ng *Engine) SendGetBalance(target *SessionData, hook Callback) {
	hops := []byte{0, 1, 2, 3, 4, 5}
	s := make(Sessions, len(hops))
	s[2] = target
	se := ng.SelectHops(hops, s)
	var c Circuit
	copy(c[:], se)
	confID := nonce.NewID()
	o := GetBalanceOnion(GetBalanceParams{target.ID, confID, se[5], c,
		ng.KeySet})
	log.D.Ln("sending out getbalance onion")
	res := ng.PostAcctOnion(o)
	ng.SendWithOneHook(c[0].AddrPort, res, hook, ng.PendingResponses)
}

func (o Skins) GetBalance(id, confID nonce.ID, prvs [3]*prv.Key,
	pubs [3]*pub.Key, nonces [3]nonce.IV) Skins {
	
	return append(o, &GetBalance{
		ID:      id,
		ConfID:  confID,
		Ciphers: GenCiphers(prvs, pubs),
		Nonces:  nonces,
		Onion:   nop,
	})
}

func (x *GetBalance) Magic() string { return GetBalanceMagic }

func (x *GetBalance) Encode(s *octet.Splice) (e error) {
	return x.Onion.Encode(s.
		Magic(GetBalanceMagic).
		ID(x.ID).ID(x.ConfID).
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
		ReadID(&x.ID).ReadID(&x.ConfID).
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
				ConfID:       x.ConfID,
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
