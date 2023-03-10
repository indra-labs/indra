package engine

import (
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
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

func (o Skins) MakeHiddenService(id nonce.ID, in *Intro,
	client *SessionData, c Circuit, ks *signer.KeySet) Skins {
	
	forwardKeys := ks.Next3()
	returnKeys := ks.Next3()
	n := GenNonces(6)
	var returnNonces, forwardNonces [3]nonce.IV
	copy(returnNonces[:], n[3:])
	copy(forwardNonces[:], n[:3])
	var forwardSessions, returnSessions [3]*SessionData
	copy(forwardSessions[:], c[:3])
	copy(returnSessions[:], c[3:5])
	returnSessions[2] = client
	var returnPubs [3]*pub.Key
	returnPubs[0] = c[3].PayloadPub
	returnPubs[1] = c[4].PayloadPub
	returnPubs[2] = client.PayloadPub
	return Skins{}.
		RoutingHeader(forwardSessions, forwardKeys, forwardNonces).
		HiddenService(id, in, returnKeys, returnPubs, returnNonces).
		RoutingHeader(returnSessions, returnKeys, returnNonces)
}

func (ng *Engine) SendHiddenService(id nonce.ID, key *prv.Key,
	target *SessionData, hook Callback) {
	
	hops := StandardCircuit()
	s := make(Sessions, len(hops))
	s[2] = target
	se := ng.SelectHops(hops, s)
	var c Circuit
	copy(c[:], se)
	in := NewIntro(key, c[2].AddrPort)
	o := Skins{}.MakeHiddenService(id, in, c[2], c, ng.KeySet)
	log.D.Ln("sending out exit onion")
	res := ng.PostAcctOnion(o)
	ng.SendWithOneHook(c[0].AddrPort, res, hook, ng.PendingResponses)
}

func (o Skins) HiddenService(id nonce.ID, in *Intro, pv [3]*prv.Key,
	pb [3]*pub.Key, nonces [3]nonce.IV) Skins {
	
	return append(o, &HiddenService{
		ID:      id,
		Intro:   *in,
		Ciphers: GenCiphers(pv, pb),
		Nonces:  nonces,
	})
}

func (x *HiddenService) Magic() string { return HiddenServiceMagic }

func (x *HiddenService) Encode(s *octet.Splice) (e error) {
	return x.Onion.Encode(s.
		Magic(HiddenServiceMagic).
		ID(x.ID).
		Pubkey(x.Key).
		AddrPort(x.AddrPort).
		Signature(&x.Sig).
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
		ReadPubkey(&x.Key).
		ReadAddrPort(&x.AddrPort).
		ReadSignature(&x.Sig).
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
