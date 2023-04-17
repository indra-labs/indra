package engine

import (
	"net/netip"
	"reflect"
	"time"
	
	"github.com/gookit/color"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/engine/coding"
	"git-indra.lan/indra-labs/indra/pkg/engine/ifc"
	"git-indra.lan/indra-labs/indra/pkg/engine/magic"
	"git-indra.lan/indra-labs/indra/pkg/engine/node"
	"git-indra.lan/indra-labs/indra/pkg/engine/sess"
	"git-indra.lan/indra-labs/indra/pkg/engine/sessions"
	"git-indra.lan/indra-labs/indra/pkg/splice"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
)

const (
	IntroMagic = "in"
	IntroLen   = magic.Len + nonce.IDLen + crypto.PubKeyLen + 1 +
		splice.AddrLen + slice.Uint64Len + crypto.SigLen
)

type Intro struct {
	ID       nonce.ID // This ensures never a repeated signed message.
	Key      *crypto.Pub
	AddrPort *netip.AddrPort
	Expiry   time.Time
	Sig      crypto.SigBytes
}

func introGen() coding.Codec           { return &Intro{} }
func init()                            { Register(IntroMagic, introGen) }
func (x *Intro) Magic() string         { return IntroMagic }
func (x *Intro) Len() int              { return IntroLen }
func (x *Intro) Wrap(inner ifc.Onion)  {}
func (x *Intro) GetOnion() interface{} { return x }

func NewIntro(id nonce.ID, key *crypto.Prv, ap *netip.AddrPort,
	expires time.Time) (in *Intro) {
	pk := crypto.DerivePub(key)
	s := splice.New(IntroLen - magic.Len)
	s.ID(id).Pubkey(pk).AddrPort(ap).Uint64(uint64(expires.
		UnixNano()))
	hash := sha256.Single(s.GetUntil(s.GetCursor()))
	var e error
	var sign crypto.SigBytes
	if sign, e = crypto.Sign(key, hash); fails(e) {
		return nil
	}
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
	s := splice.New(IntroLen - magic.Len)
	s.ID(x.ID).Pubkey(x.Key).AddrPort(x.AddrPort).Uint64(uint64(x.Expiry.
		UnixNano()))
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

func SpliceIntro(s *splice.Splice, x *Intro) *splice.Splice {
	return s.ID(x.ID).
		Pubkey(x.Key).
		AddrPort(x.AddrPort).
		Uint64(uint64(x.Expiry.UnixNano())).
		Signature(&x.Sig)
}

func (x *Intro) Encode(s *splice.Splice) (e error) {
	log.T.S("encoding", reflect.TypeOf(x),
		x.ID, x.AddrPort.String(), x.Expiry, x.Sig,
	)
	SpliceIntro(s.Magic(IntroMagic), x)
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
		ReadTime(&x.Expiry).
		ReadSignature(&x.Sig)
	return
}

func (x *Intro) Handle(s *splice.Splice, p ifc.Onion, ng ifc.Ngin) (e error) {
	ng.Hidden().Lock()
	valid := x.Validate()
	if valid {
		log.T.Ln(ng.Mgr().GetLocalNodeAddressString(), "validated intro", x.ID)
		kb := x.Key.ToBytes()
		if _, ok := ng.Hidden().KnownIntros[x.Key.ToBytes()]; ok {
			log.D.Ln(ng.Mgr().GetLocalNodeAddressString(), "already have intro")
			ng.Pending().ProcessAndDelete(x.ID, &kb, s.GetAll())
			ng.Hidden().Unlock()
			return
		}
		log.D.F("%s storing intro for %s %s",
			ng.Mgr().GetLocalNodeAddressString(), x.Key.ToBase32Abbreviated(),
			x.ID)
		ng.Hidden().KnownIntros[x.Key.ToBytes()] = x
		var ok bool
		if ok, e = ng.Pending().ProcessAndDelete(x.ID, &kb,
			s.GetAll()); ok || fails(e) {
			
			ng.Hidden().Unlock()
			log.D.Ln("deleted pending response", x.ID)
			return
		}
		log.D.F("%s sending out intro to %s at %s to all known peers",
			ng.Mgr().GetLocalNodeAddressString(), x.Key.ToBase32Abbreviated(),
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
	ng.Hidden().Unlock()
	return
}

func (x *Intro) Account(res *sess.Data, sm *sess.Manager,
	s *sessions.Data, last bool) (skip bool, sd *sessions.Data) {
	
	res.ID = x.ID
	return
}
