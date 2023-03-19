package ngin

import (
	"time"
	
	"github.com/gookit/color"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/signer"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/ngin/magic"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
	"git-indra.lan/indra-labs/indra/pkg/util/zip"
)

const (
	HiddenServiceMagic = "hs"
	HiddenServiceLen   = magic.Len + IntroLen +
		3*sha256.Len + nonce.IVLen*3 + RoutingHeaderLen
)

type HiddenService struct {
	Intro
	// Ciphers is a set of 3 symmetric ciphers that are to be used in their
	// given order over the reply message from the service.
	Ciphers [3]sha256.Hash
	// Nonces are the nonces to use with the cipher when creating the
	// encryption for the reply message,
	// they are common with the crypts in the header.
	Nonces [3]nonce.IV
	slice.Bytes
	Onion
}

func hiddenServicePrototype() Onion { return &HiddenService{} }

func init() { Register(HiddenServiceMagic, hiddenServicePrototype) }

func (o Skins) HiddenService(in *Intro, point *ExitPoint,
	header slice.Bytes) Skins {
	
	return append(o, &HiddenService{
		Intro:   *in,
		Ciphers: GenCiphers(point.Keys, point.ReturnPubs),
		Nonces:  point.Nonces,
		Bytes:   header,
		Onion:   NewTmpl(),
	})
}

func (x *HiddenService) Magic() string { return HiddenServiceMagic }

func (x *HiddenService) Encode(s *zip.Splice) (e error) {
	s.Magic(HiddenServiceMagic).
		ID(x.Intro.ID).
		Pubkey(x.Intro.Key).
		AddrPort(x.Intro.AddrPort).
		Uint64(uint64(x.Intro.Expiry.UnixNano())).
		Signature(&x.Intro.Sig).
		HashTriple(x.Ciphers).
		IVTriple(x.Nonces).
		RoutingHeader(x.Bytes)
	return
}

func (x *HiddenService) Decode(s *zip.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), HiddenServiceLen-magic.Len,
		HiddenServiceMagic); check(e) {
		return
	}
	s.ReadID(&x.Intro.ID).
		ReadPubkey(&x.Intro.Key).
		ReadAddrPort(&x.Intro.AddrPort).
		ReadTime(&x.Intro.Expiry).
		ReadSignature(&x.Intro.Sig).
		ReadHashTriple(&x.Ciphers).
		ReadIVTriple(&x.Nonces).
		ReadRoutingHeader(&x.Bytes)
	return
}

func (x *HiddenService) Len() int { return HiddenServiceLen + x.Onion.Len() }

func (x *HiddenService) Wrap(inner Onion) { x.Onion = inner }

func (x *HiddenService) Handle(s *zip.Splice, p Onion, ng *Engine) (e error) {
	log.D.F("%s adding introduction for key %s",
		ng.GetLocalNodeAddressString(), x.Key.ToBase32Abbreviated())
	ng.HiddenRouting.AddIntro(x.Key, &Introduction{
		Intro:   &x.Intro,
		Ciphers: x.Ciphers,
		Nonces:  x.Nonces,
		Bytes:   s.GetCursorToEnd(),
	})
	log.D.S(ng.GetLocalNodeAddressString(), ng.HiddenRouting)
	log.D.Ln("stored new introduction, starting broadcast")
	go GossipIntro(&x.Intro, ng.SessionManager, ng.C)
	return
}

func MakeHiddenService(header slice.Bytes, in *Intro, alice, bob *SessionData,
	c Circuit, ks *signer.KeySet) Skins {
	
	headers := GetHeaders(alice, bob, c, ks)
	return Skins{}.
		RoutingHeader(headers.Forward).
		HiddenService(in, headers.ExitPoint(), header).
		RoutingHeader(headers.Return)
}

func (ng *Engine) SendHiddenService(
	header slice.Bytes,
	id nonce.ID,
	key *prv.Key,
	expiry time.Time,
	alice, bob *SessionData,
	localPort uint16,
	hook Callback) {
	
	hops := StandardCircuit()
	s := make(Sessions, len(hops))
	s[2] = alice
	se := ng.SelectHops(hops, s)
	var c Circuit
	copy(c[:], se[:len(c)])
	in := NewIntro(id, key, alice.AddrPort, expiry)
	log.D.Ln("intro", in, in.Validate())
	o := MakeHiddenService(header, in, alice, bob, c, ng.KeySet)
	log.D.F("%s sending out hidden service onion %s",
		ng.GetLocalNodeAddressString(),
		color.Yellow.Sprint(alice.AddrPort.String()))
	res := ng.PostAcctOnion(o)
	log.D.Ln("storing hidden service info")
	ng.HiddenRouting.AddHiddenService(key, localPort, ng.GetLocalNodeAddressString())
	ng.SendWithOneHook(c[0].AddrPort, res, hook, ng.PendingResponses)
}
