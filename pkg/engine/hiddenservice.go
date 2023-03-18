package engine

import (
	"time"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/signer"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/engine/magic"
	"git-indra.lan/indra-labs/indra/pkg/util/zip"
)

const (
	HiddenServiceMagic = "hs"
	HiddenServiceLen   = magic.Len + nonce.IDLen + IntroLen +
		3*sha256.Len + nonce.IVLen*3
)

type HiddenService struct {
	Intro
	zip.Reply
	Onion
}

func hiddenServicePrototype() Onion { return &HiddenService{} }

func init() { Register(HiddenServiceMagic, hiddenServicePrototype) }

func (o Skins) HiddenService(in *Intro, point *ExitPoint) Skins {
	return append(o, &HiddenService{
		Intro: *in,
		Reply: zip.Reply{
			ID:      in.ID,
			Ciphers: GenCiphers(point.Keys, point.ReturnPubs),
			Nonces:  point.Nonces,
		},
		Onion: NewTmpl(),
	})
}

func (x *HiddenService) Magic() string { return HiddenServiceMagic }

func (x *HiddenService) Encode(s *zip.Splice) (e error) {
	return x.Onion.Encode(s.
		Magic(HiddenServiceMagic).
		ID(x.Intro.ID).
		Pubkey(x.Intro.Key).
		AddrPort(x.Intro.AddrPort).
		Uint64(uint64(x.Intro.Expiry.UnixNano())).
		Signature(&x.Intro.Sig).
		Reply(&x.Reply),
	)
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
		ReadReply(&x.Reply)
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
	log.D.Ln("stored new introduction, starting broadcast")
	go GossipIntro(&x.Intro, ng.SessionManager, ng.C)
	return
}

func MakeHiddenService(in *Intro, client *SessionData, c Circuit,
	ks *signer.KeySet) Skins {
	
	headers := GetHeaders(client, c, ks)
	return Skins{}.
		RoutingHeader(headers.Forward).
		HiddenService(in, headers.ExitPoint()).
		RoutingHeader(headers.Return)
}

func (ng *Engine) SendHiddenService(id nonce.ID, key *prv.Key, expiry time.Time,
	target *SessionData, localPort uint16, hook Callback) {
	
	hops := StandardCircuit()
	s := make(Sessions, len(hops))
	se := ng.SelectHops(hops, s)
	var c Circuit
	copy(c[:], se[:len(c)])
	in := NewIntro(id, key, target.AddrPort, expiry)
	log.D.Ln("intro", in, in.Validate())
	o := MakeHiddenService(in, target, c, ng.KeySet)
	log.I.S("sending out hidden service onion")
	res := ng.PostAcctOnion(o)
	log.D.Ln("storing hidden service info")
	ng.HiddenRouting.AddHiddenService(key, localPort)
	ng.SendWithOneHook(c[0].AddrPort, res, hook, ng.PendingResponses)
}
