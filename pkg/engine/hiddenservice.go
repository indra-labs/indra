package engine

import (
	"time"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/signer"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/engine/magic"
	"git-indra.lan/indra-labs/indra/pkg/util/octet"
)

const (
	HiddenServiceMagic = "hs"
	HiddenServiceLen   = magic.Len + nonce.IDLen + IntroLen +
		3*sha256.Len + nonce.IVLen*3
)

type HiddenService struct {
	nonce.ID
	Intro
	// Ciphers is a set of 3 symmetric ciphers that are to be used in their
	// given order over the reply message from the service.
	Ciphers [3]sha256.Hash
	// Nonces are the nonces to use with the cipher when creating the
	// encryption for the reply message.
	Nonces [3]nonce.IV
	Onion
}

func hiddenServicePrototype() Onion { return &HiddenService{} }

func init() { Register(HiddenServiceMagic, hiddenServicePrototype) }

func MakeHiddenService(id nonce.ID, in *Intro,
	client *SessionData, c Circuit, ks *signer.KeySet) Skins {
	
	headers := GetHeaders(client, c, ks)
	return Skins{}.
		RoutingHeader(headers.Forward).
		HiddenService(id, in, headers.ExitPoint()).
		RoutingHeader(headers.Return)
}

func (ng *Engine) SendHiddenService(id nonce.ID, key *prv.Key, expiry time.Time,
	target *SessionData, hook Callback) {
	
	hops := StandardCircuit()
	s := make(Sessions, len(hops))
	s[2] = target
	se := ng.SelectHops(hops, s)
	var c Circuit
	copy(c[:], se)
	in := NewIntro(id, key, c[2].AddrPort, expiry)
	log.D.Ln("intro", in, in.Validate())
	o := MakeHiddenService(id, in, c[2], c, ng.KeySet)
	log.D.Ln("sending out hidden service onion")
	res := ng.PostAcctOnion(o)
	ng.SendWithOneHook(c[0].AddrPort, res, hook, ng.PendingResponses)
}

func (o Skins) HiddenService(id nonce.ID, in *Intro, point *ExitPoint) Skins {
	
	return append(o, &HiddenService{
		ID:      id,
		Intro:   *in,
		Ciphers: GenCiphers(point.Keys, point.ReturnPubs),
		Nonces:  point.Nonces,
	})
}

func (x *HiddenService) Magic() string { return HiddenServiceMagic }

func (x *HiddenService) Encode(s *octet.Splice) (e error) {
	return x.Onion.Encode(s.
		Magic(HiddenServiceMagic).
		ID(x.ID).
		ID(x.Intro.ID).
		Pubkey(x.Intro.Key).
		AddrPort(x.Intro.AddrPort).
		Uint64(uint64(x.Intro.Expiry.UnixNano())).
		Signature(&x.Intro.Sig).
		HashTriple(x.Ciphers).
		IVTriple(x.Nonces),
	)
}

func (x *HiddenService) Decode(s *octet.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), HiddenServiceLen-magic.Len,
		HiddenServiceMagic); check(e) {
		return
	}
	s.
		ReadID(&x.ID).
		ReadID(&x.Intro.ID).
		ReadPubkey(&x.Intro.Key).
		ReadAddrPort(&x.Intro.AddrPort).
		ReadTime(&x.Intro.Expiry).
		ReadSignature(&x.Intro.Sig).
		ReadHashTriple(&x.Ciphers).
		ReadIVTriple(&x.Nonces)
	return
}

func (x *HiddenService) Len() int { return HiddenServiceLen + x.Onion.Len() }

func (x *HiddenService) Wrap(inner Onion) { x.Onion = inner }

func (x *HiddenService) Handle(s *octet.Splice, p Onion, ng *Engine) (e error) {
	log.D.F("%s adding introduction for key %s",
		ng.GetLocalNodeAddress(), x.Key.ToBase32())
	ng.Introductions.AddIntro(x.Key, s.GetCursorToEnd())
	log.D.Ln("stored new introduction, starting broadcast")
	go GossipIntro(&x.Intro, ng.SessionManager, ng.C)
	return
}

/*
 
 */
