package onions

import (
	"net/netip"
	"reflect"
	"time"
	
	"git-indra.lan/indra-labs/indra/pkg/crypto"
	"git-indra.lan/indra-labs/indra/pkg/crypto/nonce"
	"git-indra.lan/indra-labs/indra/pkg/crypto/sha256"
	"git-indra.lan/indra-labs/indra/pkg/engine/coding"
	"git-indra.lan/indra-labs/indra/pkg/engine/magic"
	"git-indra.lan/indra-labs/indra/pkg/engine/sess"
	"git-indra.lan/indra-labs/indra/pkg/engine/sessions"
	"git-indra.lan/indra-labs/indra/pkg/util/slice"
	"git-indra.lan/indra-labs/indra/pkg/util/splice"
)

const (
	IntroMagic = "intr"
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
func (x *Intro) Wrap(inner Onion)      {}
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

func (x *Intro) Handle(s *splice.Splice, p Onion, ng Ngin) (e error) {
	ng.GetHidden().Lock()
	valid := x.Validate()
	if valid {
		log.T.Ln(ng.Mgr().GetLocalNodeAddressString(), "validated intro", x.ID)
	}
	ng.GetHidden().Unlock()
	return
}

func (x *Intro) Account(res *sess.Data, sm *sess.Manager,
	s *sessions.Data, last bool) (skip bool, sd *sessions.Data) {
	
	res.ID = x.ID
	return
}
