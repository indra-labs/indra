package engine

import (
	"net/netip"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/sig"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/engine/magic"
	"git-indra.lan/indra-labs/indra/pkg/util/octet"
)

const (
	IntroMagic = "in"
	IntroLen   = magic.Len + pub.KeyLen + 1 + octet.AddrLen + sig.Len
)

type Intro struct {
	Key      *pub.Key
	AddrPort *netip.AddrPort
	Sig      sig.Bytes
}

func introPrototype() Onion { return &Intro{} }

func init() { Register(IntroMagic, introPrototype) }

func (o Skins) Intro(key *prv.Key, ap *netip.AddrPort) (sk Skins) {
	return append(o, NewIntro(key, ap))
}

func NewIntro(key *prv.Key, ap *netip.AddrPort) (in *Intro) {
	pk := pub.Derive(key)
	bap, _ := ap.MarshalBinary()
	pkb := pk.ToBytes()
	hash := sha256.Single(append(pkb[:], bap...))
	var e error
	var sign sig.Bytes
	if sign, e = sig.Sign(key, hash); check(e) {
		return nil
	}
	in = &Intro{
		Key:      pk,
		AddrPort: ap,
		Sig:      sign,
	}
	return
}

func (x *Intro) Validate() bool {
	bap, _ := x.AddrPort.MarshalBinary()
	pkb := x.Key.ToBytes()
	hash := sha256.Single(append(pkb[:], bap...))
	key, e := x.Sig.Recover(hash)
	if check(e) {
		return false
	}
	kb := key.ToBytes()
	if kb.Equals(pkb) {
		return true
	}
	return false
}

func (x *Intro) Magic() string { return IntroMagic }

func (x *Intro) Encode(s *octet.Splice) (e error) {
	s.
		Magic(IntroMagic).
		Pubkey(x.Key).
		AddrPort(x.AddrPort).
		Signature(&x.Sig)
	return
}

func (x *Intro) Decode(s *octet.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), IntroLen-magic.Len, IntroMagic); check(e) {
		return
	}
	s.
		ReadPubkey(&x.Key).
		ReadAddrPort(&x.AddrPort).
		ReadSignature(&x.Sig)
	return
}

func (x *Intro) Len() int { return IntroLen }

func (x *Intro) Wrap(inner Onion) {}

func (x *Intro) Handle(s *octet.Splice, p Onion,
	ng *Engine) (e error) {
	
	ng.Introductions.Lock()
	if x.Validate() {
		if _, ok := ng.Introductions.KnownIntros[x.Key.ToBytes()]; ok {
			ng.Introductions.Unlock()
			return
		}
		log.D.F("%s storing intro for %s",
			ng.GetLocalNodeAddress().String(), x.Key.ToBase32())
		ng.Introductions.KnownIntros[x.Key.ToBytes()] = x
		log.D.F("%s sending out intro to %s at %s to all known peers",
			ng.GetLocalNodeAddress(), x.Key.ToBase32(), x.AddrPort.String())
		sender := ng.SessionManager.FindNodeByAddrPort(x.AddrPort)
		nn := make(map[nonce.ID]*Node)
		ng.SessionManager.ForEachNode(func(n *Node) bool {
			if n.ID != sender.ID {
				nn[n.ID] = n
			}
			return false
		})
		counter := 0
		for i := range nn {
			log.T.F("sending intro to %s", nn[i].AddrPort.String())
			nn[i].Transport.Send(s.GetRange(-1, -1))
			counter++
			if counter < 2 {
				continue
			}
			break
		}
		ng.Introductions.Unlock()
	}
	return
}
