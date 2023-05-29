package onions

import (
	"net/netip"
	"reflect"
	"time"

	"github.com/gookit/color"

	"git-indra.lan/indra-labs/indra/pkg/crypto"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/engine/coding"
	"git-indra.lan/indra-labs/indra/pkg/engine/magic"
	"git-indra.lan/indra-labs/indra/pkg/engine/node"
	"git-indra.lan/indra-labs/indra/pkg/engine/sess"
	"git-indra.lan/indra-labs/indra/pkg/engine/sessions"
	"git-indra.lan/indra-labs/indra/pkg/util/qu"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
	"git-indra.lan/indra-labs/indra/pkg/util/splice"
)

const (
	IntroMagic = "intr"
	IntroLen   = magic.Len +
		nonce.IDLen +
		crypto.PubKeyLen + 1 + 1 +
		splice.AddrLen +
		slice.Uint16Len +
		slice.Uint32Len +
		slice.Uint64Len +
		crypto.SigLen
)

type Intro struct {
	ID        nonce.ID        // Ensures never a repeated signature.
	Key       *crypto.Pub     // Hidden service address.
	AddrPort  *netip.AddrPort // Introducer address.
	Port      uint16          // Well known port of protocol available.
	RelayRate uint32          // mSat/Mb
	Expiry    time.Time
	Sig       crypto.SigBytes
}

func introGen() coding.Codec           { return &Intro{} }
func init()                            { Register(IntroMagic, introGen) }
func (x *Intro) Magic() string         { return IntroMagic }
func (x *Intro) Len() int              { return IntroLen }
func (x *Intro) Wrap(inner Onion)      {}
func (x *Intro) GetOnion() interface{} { return x }

func NewIntro(id nonce.ID, key *crypto.Prv, ap *netip.AddrPort,
	relayRate uint32, port uint16, expires time.Time) (in *Intro) {

	pk := crypto.DerivePub(key)
	s := splice.New(IntroLen - magic.Len)
	s.ID(id).
		Pubkey(pk).
		AddrPort(ap).
		Uint32(relayRate).
		Uint16(port).
		Uint64(uint64(expires.UnixNano()))
	hash := sha256.Single(s.GetUntil(s.GetCursor()))
	var e error
	var sign crypto.SigBytes
	if sign, e = crypto.Sign(key, hash); fails(e) {
		return nil
	}
	in = &Intro{
		ID:        id,
		Key:       pk,
		AddrPort:  ap,
		RelayRate: relayRate,
		Port:      port,
		Expiry:    expires,
		Sig:       sign,
	}
	return
}

func (x *Intro) Validate() bool {
	s := splice.New(IntroLen - magic.Len)
	s.ID(x.ID).
		Pubkey(x.Key).
		AddrPort(x.AddrPort).
		Uint32(x.RelayRate).
		Uint16(x.Port).
		Uint64(uint64(x.Expiry.UnixNano()))
	hash := sha256.Single(s.GetUntil(s.GetCursor()))
	key, e := x.Sig.Recover(hash)
	if fails(e) {
		return false
	}
	if key.Equals(x.Key) && x.Expiry.After(time.Now()) {
		return true
	}
	return false
}

func (x *Intro) Splice(s *splice.Splice) {
	s.ID(x.ID).
		Pubkey(x.Key).
		AddrPort(x.AddrPort).
		Uint32(x.RelayRate).
		Uint16(x.Port).
		Uint64(uint64(x.Expiry.UnixNano())).
		Signature(x.Sig)
}

func (x *Intro) Encode(s *splice.Splice) (e error) {
	log.T.S("encoding", reflect.TypeOf(x),
		x.ID, x.AddrPort.String(), x.Expiry, x.Sig,
	)
	x.Splice(s.Magic(IntroMagic))
	return
}

func (x *Intro) Decode(s *splice.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), IntroLen-magic.Len,
		IntroMagic); fails(e) {

		return
	}
	s.
		ReadID(&x.ID).
		ReadPubkey(&x.Key).
		ReadAddrPort(&x.AddrPort).
		ReadUint32(&x.RelayRate).
		ReadUint16(&x.Port).
		ReadTime(&x.Expiry).
		ReadSignature(&x.Sig)
	return
}

func (x *Intro) Handle(s *splice.Splice, p Onion, ng Ngin) (e error) {
	ng.GetHidden().Lock()
	valid := x.Validate()
	if valid {
		log.T.Ln(ng.Mgr().GetLocalNodeAddressString(), "validated intro", x.ID)
		kb := x.Key.ToBytes()
		if _, ok := ng.GetHidden().KnownIntros[x.Key.ToBytes()]; ok {
			log.D.Ln(ng.Mgr().GetLocalNodeAddressString(), "already have intro")
			ng.Pending().ProcessAndDelete(x.ID, &kb, s.GetAll())
			ng.GetHidden().Unlock()
			return
		}
		log.D.F("%s storing intro for %s %s",
			ng.Mgr().GetLocalNodeAddressString(), x.Key.ToBased32Abbreviated(),
			x.ID)
		ng.GetHidden().KnownIntros[x.Key.ToBytes()] = x
		var ok bool
		if ok, e = ng.Pending().ProcessAndDelete(x.ID, &kb,
			s.GetAll()); ok || fails(e) {

			ng.GetHidden().Unlock()
			log.D.Ln("deleted pending response", x.ID)
			return
		}
		log.D.F("%s sending out intro to %s at %s to all known peers",
			ng.Mgr().GetLocalNodeAddressString(), x.Key.ToBased32Abbreviated(),
			color.Yellow.Sprint(x.AddrPort.String()))
		sender := ng.Mgr().FindNodeByAddrPort(x.AddrPort)
		nn := make(map[nonce.ID]*node.Node)
		ng.Mgr().ForEachNode(func(n *node.Node) bool {
			if n.ID != sender.ID {
				nn[n.ID] = n
				return true
			}
			return false
		})
		counter := 0
		for i := range nn {
			log.T.F("sending intro to %s", color.Yellow.Sprint(nn[i].AddrPort.
				String()))
			nn[i].Transport.Send(s.GetAll())
			counter++
			if counter < 2 {
				continue
			}
			break
		}
	}
	ng.GetHidden().Unlock()
	return
}

func (x *Intro) Account(res *sess.Data, sm *sess.Manager,
	s *sessions.Data, last bool) (skip bool, sd *sessions.Data) {

	res.ID = x.ID
	return
}

func (x *Intro) Gossip(sm *sess.Manager, c qu.C) {
	log.D.F("propagating hidden service intro for %s",
		x.Key.ToBased32Abbreviated())
	Gossip(x, sm, c)
	log.T.Ln("finished broadcasting intro")
}
