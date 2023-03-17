package engine

import (
	"net/netip"
	"time"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/prv"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/pub"
	"git-indra.lan/indra-labs/indra/pkg/crypto/key/sig"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/engine/magic"
	"git-indra.lan/indra-labs/indra/pkg/util/octet"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

const (
	IntroMagic = "in"
	IntroLen   = magic.Len + nonce.IDLen + pub.KeyLen + 1 +
		octet.AddrLen + slice.Uint64Len + sig.Len
)

type Intro struct {
	nonce.ID // This ensures never a repeated signed message.
	Key      *pub.Key
	AddrPort *netip.AddrPort
	Expiry   time.Time
	Sig      sig.Bytes
}

func introPrototype() Onion { return &Intro{} }

func init() { Register(IntroMagic, introPrototype) }

func (o Skins) Intro(id nonce.ID, key *prv.Key, ap *netip.AddrPort,
	expires time.Time) (sk Skins) {
	return append(o, NewIntro(id, key, ap, expires))
}

func NewIntro(id nonce.ID, key *prv.Key, ap *netip.AddrPort,
	expires time.Time) (in *Intro) {
	pk := pub.Derive(key)
	s := octet.New(IntroLen - magic.Len)
	s.ID(id).Pubkey(pk).AddrPort(ap).Uint64(uint64(expires.
		UnixNano()))
	hash := sha256.Single(s.GetRange(-1, s.GetCursor()))
	var e error
	var sign sig.Bytes
	if sign, e = sig.Sign(key, hash); check(e) {
		return nil
	}
	log.T.S("new intro bytes", s.GetRange(-1, s.GetCursor()).ToBytes(),
		sign)
	in = &Intro{
		ID:       id,
		Key:      pk,
		AddrPort: ap,
		Expiry:   expires,
		Sig:      sign,
	}
	return
}

func (x *Intro) Validate() bool {
	s := octet.New(IntroLen - magic.Len)
	s.ID(x.ID).Pubkey(x.Key).AddrPort(x.AddrPort).Uint64(uint64(x.Expiry.
		UnixNano()))
	hash := sha256.Single(s.GetRange(-1, s.GetCursor()))
	key, e := x.Sig.Recover(hash)
	if check(e) {
		return false
	}
	if key.Equals(x.Key) && x.Expiry.After(time.Now()) {
		return true
	}
	return false
}

func (x *Intro) Magic() string { return IntroMagic }

func (x *Intro) Encode(s *octet.Splice) (e error) {
	s.Magic(IntroMagic).
		ID(x.ID).
		Pubkey(x.Key).
		AddrPort(x.AddrPort).
		Uint64(uint64(x.Expiry.UnixNano())).
		Signature(&x.Sig)
	return
}

func (x *Intro) Decode(s *octet.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), IntroLen-magic.Len,
		IntroMagic); check(e) {
		
		return
	}
	s.ReadID(&x.ID).
		ReadPubkey(&x.Key).
		ReadAddrPort(&x.AddrPort).
		ReadTime(&x.Expiry).
		ReadSignature(&x.Sig)
	return
}

func (x *Intro) Len() int { return IntroLen }

func (x *Intro) Wrap(inner Onion) {}

func (x *Intro) Handle(s *octet.Splice, p Onion,
	ng *Engine) (e error) {
	
	ng.Introductions.Lock()
	valid := x.Validate()
	if valid {
		log.T.Ln(ng.GetLocalNodeAddress().String(), "validated intro", x.ID)
		// ng.PendingResponses.ProcessAndDelete(x.ID, s.GetRange(-1, -1))
		kb := x.Key.ToBytes()
		if _, ok := ng.Introductions.KnownIntros[x.Key.ToBytes()]; ok {
			log.D.Ln(ng.GetLocalNodeAddress(), "already have intro")
			ng.PendingResponses.ProcessAndDelete(x.ID, &kb, s.GetRange(-1, -1))
			ng.Introductions.Unlock()
			return
		}
		log.D.F("%s storing intro for %s %s",
			ng.GetLocalNodeAddress().String(), x.Key.ToBase32Abbreviated(), x.ID)
		ng.Introductions.KnownIntros[x.Key.ToBytes()] = x
		var ok bool
		if ok, e = ng.PendingResponses.ProcessAndDelete(x.ID, &kb,
			s.GetRange(-1, -1)); ok || check(e) {
			
			ng.Introductions.Unlock()
			log.D.Ln("deleted pending response", x.ID)
			return
		}
		log.D.F("%s sending out intro to %s at %s to all known peers",
			ng.GetLocalNodeAddress(), x.Key.ToBase32Abbreviated(),
			x.AddrPort.String())
		sender := ng.SessionManager.FindNodeByAddrPort(x.AddrPort)
		nn := make(map[nonce.ID]*Node)
		ng.SessionManager.ForEachNode(func(n *Node) bool {
			if n.ID != sender.ID {
				nn[n.ID] = n
				return true
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
	}
	ng.Introductions.Unlock()
	return
}
