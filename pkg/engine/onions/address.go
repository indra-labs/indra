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
	AddrMagic = "addr"
	AddrLen   = magic.Len +
		nonce.IDLen +
		crypto.PubKeyLen + 1 +
		splice.AddrLen +
		slice.Uint64Len +
		crypto.SigLen
)

type Addr struct {
	ID       nonce.ID        // This ensures never a repeated signed message.
	Key      *crypto.Pub     // Identity key.
	AddrPort *netip.AddrPort // Introducer address.
	Expiry   time.Time
	Sig      crypto.SigBytes
}

func addrGen() coding.Codec           { return &Addr{} }
func init()                           { Register(AddrMagic, addrGen) }
func (x *Addr) Magic() string         { return AddrMagic }
func (x *Addr) Len() int              { return AddrLen }
func (x *Addr) Wrap(inner Onion)      {}
func (x *Addr) GetOnion() interface{} { return x }

func NewAddr(id nonce.ID, key *crypto.Prv, addr *netip.AddrPort,
	expires time.Time) (in *Addr) {

	pk := crypto.DerivePub(key)
	s := splice.New(AddrLen - magic.Len)
	s.ID(id).
		Pubkey(pk).
		AddrPort(addr).
		Uint64(uint64(expires.UnixNano()))
	hash := sha256.Single(s.GetUntil(s.GetCursor()))
	var e error
	var sign crypto.SigBytes
	if sign, e = crypto.Sign(key, hash); fails(e) {
		return nil
	}
	in = &Addr{
		ID:       id,
		Key:      pk,
		AddrPort: addr,
		Expiry:   expires,
		Sig:      sign,
	}
	return
}

func (x *Addr) Validate() bool {
	s := splice.New(AddrLen - magic.Len)
	s.ID(x.ID).
		Pubkey(x.Key).
		AddrPort(x.AddrPort).
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

func (x *Addr) Splice(s *splice.Splice) {
	s.ID(x.ID).
		Pubkey(x.Key).
		AddrPort(x.AddrPort).
		Uint64(uint64(x.Expiry.UnixNano())).
		Signature(x.Sig)
}

func (x *Addr) Encode(s *splice.Splice) (e error) {
	log.T.S("encoding", reflect.TypeOf(x),
		x.ID, x.Expiry, x.Sig,
	)
	x.Splice(s.Magic(AddrMagic))
	return
}

func (x *Addr) Decode(s *splice.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), AddrLen-magic.Len,
		AddrMagic); fails(e) {

		return
	}
	s.ReadID(&x.ID).
		ReadPubkey(&x.Key).
		ReadAddrPort(&x.AddrPort).
		ReadTime(&x.Expiry).
		ReadSignature(&x.Sig)
	return
}

func (x *Addr) Handle(s *splice.Splice, p Onion, ng Ngin) (e error) {
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
			ng.Mgr().GetLocalNodeAddressString(), x.Key.ToBase32Abbreviated(),
			x.ID)
		// ng.GetHidden().KnownIntros[x.Key.ToBytes()] = x
		var ok bool
		if ok, e = ng.Pending().ProcessAndDelete(x.ID, &kb,
			s.GetAll()); ok || fails(e) {

			ng.GetHidden().Unlock()
			log.D.Ln("deleted pending response", x.ID)
			return
		}
		log.D.F("%s sending out intro to %s to all known peers",
			ng.Mgr().GetLocalNodeAddressString(), x.Key.ToBase32Abbreviated())
		sender := ng.Mgr().FindNodeByIdentity(x.Key)
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

func (x *Addr) Account(res *sess.Data, sm *sess.Manager,
	s *sessions.Data, last bool) (skip bool, sd *sessions.Data) {

	res.ID = x.ID
	return
}

func (x *Addr) Gossip(sm *sess.Manager, c qu.C) {
	log.D.F("propagating peer info for %s",
		x.Key.ToBase32Abbreviated())
	Gossip(x, sm, c)
	log.T.Ln("finished broadcasting peer info")
}
