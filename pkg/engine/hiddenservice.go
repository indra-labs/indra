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

func (o Skins) MakeHiddenService(id nonce.ID, il *Intro,
	client *SessionData, s Circuit, ks *signer.KeySet) Skins {
	
	var prvs [3]*prv.Key
	for i := range prvs {
		prvs[i] = ks.Next()
	}
	n := GenNonces(6)
	var returnNonces [3]nonce.IV
	copy(returnNonces[:], n[3:])
	var pubs [3]*pub.Key
	pubs[0] = s[3].PayloadPub
	pubs[1] = s[4].PayloadPub
	pubs[2] = client.PayloadPub
	return o.
		ReverseCrypt(s[0], ks.Next(), n[0], 3).
		ReverseCrypt(s[1], ks.Next(), n[1], 2).
		ReverseCrypt(s[2], ks.Next(), n[2], 1).
		HiddenService(id, il, prvs, pubs, returnNonces).
		ReverseCrypt(s[3], prvs[0], n[3], 3).
		ReverseCrypt(s[4], prvs[1], n[4], 2).
		ReverseCrypt(client, prvs[2], n[5], 1)
}

func (o Skins) HiddenService(id nonce.ID, intr *Intro, prvs [3]*prv.Key,
	pubs [3]*pub.Key, nonces [3]nonce.IV) Skins {
	
	return append(o, &HiddenService{
		ID:      id,
		Intro:   *intr,
		Ciphers: GenCiphers(prvs, pubs),
		Nonces:  nonces,
	})
}

func NewHiddenService() *HiddenService {
	return &HiddenService{}
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
	
	log.D.F("%s adding introduction for key %s", ng.GetLocalNodeAddress(),
		x.Key.ToBase32())
	ng.Introductions.AddIntro(x.Key, s.GetRange(s.GetCursor(), -1))
	log.I.Ln("stored new introduction, starting broadcast")
	go GossipIntro(&x.Intro, ng.SessionManager, ng.C)
	return
}
