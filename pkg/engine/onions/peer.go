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
	PeerMagic = "peer"
	PeerLen   = magic.Len +
		nonce.IDLen +
		crypto.PubKeyLen + 1 +
		splice.AddrLen +
		slice.Uint32Len +
		slice.Uint64Len +
		crypto.SigLen
)

type Peer struct {
	ID        nonce.ID    // This ensures never a repeated signed message.
	Key       *crypto.Pub // Identity key.
	AddrPort  *netip.AddrPort
	RelayRate uint32
	Expiry    time.Time
	Sig       crypto.SigBytes
}

func NewPeer(id nonce.ID, key *crypto.Prv, expires time.Time,
	relayRate uint32) (in *Peer) {

	pk := crypto.DerivePub(key)
	s := splice.New(IntroLen - magic.Len)
	s.ID(id).
		Pubkey(pk).
		Uint32(relayRate).
		Uint64(uint64(expires.UnixNano()))
	hash := sha256.Single(s.GetUntil(s.GetCursor()))
	var e error
	var sign crypto.SigBytes
	if sign, e = crypto.Sign(key, hash); fails(e) {
		return nil
	}
	in = &Peer{
		ID:        id,
		Key:       pk,
		RelayRate: relayRate,
		Expiry:    expires,
		Sig:       sign,
	}
	return
}

func (x *Peer) Account(res *sess.Data, sm *sess.Manager,
	s *sessions.Data, last bool) (skip bool, sd *sessions.Data) {

	res.ID = x.ID
	return
}

func (x *Peer) Decode(s *splice.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), PeerLen-magic.Len,
		PeerMagic); fails(e) {

		return
	}
	s.ReadID(&x.ID).
		ReadPubkey(&x.Key).
		ReadUint32(&x.RelayRate).
		ReadTime(&x.Expiry).
		ReadSignature(&x.Sig)
	return
}

func (x *Peer) Encode(s *splice.Splice) (e error) {
	log.T.S("encoding", reflect.TypeOf(x),
		x.ID, x.Expiry, x.Sig,
	)
	x.Splice(s.Magic(PeerMagic))
	return
}

func (x *Peer) GetOnion() interface{} { return x }

func (x *Peer) Gossip(sm *sess.Manager, c qu.C) {
	log.D.F("propagating peer info for %s",
		x.Key.ToBased32Abbreviated())
	Gossip(x, sm, c)
	log.T.Ln("finished broadcasting peer info")
}

func (x *Peer) Handle(s *splice.Splice, p Onion, ng Ngin) (e error) {
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
		// ng.GetHidden().KnownIntros[x.Key.ToBytes()] = x
		var ok bool
		if ok, e = ng.Pending().ProcessAndDelete(x.ID, &kb,
			s.GetAll()); ok || fails(e) {

			ng.GetHidden().Unlock()
			log.D.Ln("deleted pending response", x.ID)
			return
		}
		log.D.F("%s sending out intro to %s to all known peers",
			ng.Mgr().GetLocalNodeAddressString(), x.Key.ToBased32Abbreviated())
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
			log.T.F("sending intro to %s",
				color.Yellow.Sprint(nn[i].AddrPort.String()))
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

func (x *Peer) Len() int      { return PeerLen }
func (x *Peer) Magic() string { return PeerMagic }

func (x *Peer) Splice(s *splice.Splice) {
	s.ID(x.ID).
		Pubkey(x.Key).
		Uint32(x.RelayRate).
		Uint64(uint64(x.Expiry.UnixNano())).
		Signature(x.Sig)
}

func (x *Peer) Validate() bool {
	s := splice.New(PeerLen - magic.Len)
	s.ID(x.ID).
		Pubkey(x.Key).
		Uint32(x.RelayRate).
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

func (x *Peer) Wrap(inner Onion) {}
func init()                      { Register(PeerMagic, peerGen) }
func peerGen() coding.Codec      { return &Peer{} }
