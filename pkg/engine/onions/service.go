package onions

import (
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
	ServiceMagic = "serv"
	ServiceLen   = magic.Len + nonce.IDLen + crypto.PubKeyLen + 1 +
		splice.AddrLen + slice.Uint64Len + crypto.SigLen
)

type Service struct {
	ID        nonce.ID    // This ensures never a repeated signed message.
	Key       *crypto.Pub // Identity key.
	RelayRate uint64
	Expiry    time.Time
	Sig       crypto.SigBytes
}

func servGen() coding.Codec              { return &Service{} }
func init()                              { Register(ServiceMagic, servGen) }
func (x *Service) Magic() string         { return ServiceMagic }
func (x *Service) Len() int              { return ServiceLen }
func (x *Service) Wrap(inner Onion)      {}
func (x *Service) GetOnion() interface{} { return x }

func NewService(id nonce.ID, key *crypto.Prv,
	expires time.Time) (in *Service) {

	pk := crypto.DerivePub(key)
	s := splice.New(ServiceLen - magic.Len)
	s.ID(id).Pubkey(pk).Uint64(uint64(expires.
		UnixNano()))
	hash := sha256.Single(s.GetUntil(s.GetCursor()))
	var e error
	var sign crypto.SigBytes
	if sign, e = crypto.Sign(key, hash); fails(e) {
		return nil
	}
	in = &Service{
		ID:     id,
		Key:    pk,
		Expiry: expires,
		Sig:    sign,
	}
	return
}

func (x *Service) Validate() bool {
	s := splice.New(ServiceLen - magic.Len)
	s.ID(x.ID).Pubkey(x.Key).Uint64(uint64(x.Expiry.
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

func (x *Service) Splice(s *splice.Splice) {
	s.ID(x.ID).
		Pubkey(x.Key).
		Uint64(uint64(x.Expiry.UnixNano())).
		Signature(x.Sig)
}

func (x *Service) Encode(s *splice.Splice) (e error) {
	log.T.S("encoding", reflect.TypeOf(x),
		x.ID, x.Expiry, x.Sig,
	)
	x.Splice(s.Magic(ServiceMagic))
	return
}

func (x *Service) Decode(s *splice.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), ServiceLen-magic.Len,
		ServiceMagic); fails(e) {

		return
	}
	s.ReadID(&x.ID).
		ReadPubkey(&x.Key).
		ReadTime(&x.Expiry).
		ReadSignature(&x.Sig)
	return
}

func (x *Service) Handle(s *splice.Splice, p Onion, ng Ngin) (e error) {
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

func (x *Service) Account(res *sess.Data, sm *sess.Manager,
	s *sessions.Data, last bool) (skip bool, sd *sessions.Data) {

	res.ID = x.ID
	return
}

func (x *Service) Gossip(sm *sess.Manager, c qu.C) {
	log.D.F("propagating peer info for %s",
		x.Key.ToBase32Abbreviated())
	Gossip(x, sm, c)
	log.T.Ln("finished broadcasting peer info")
}
