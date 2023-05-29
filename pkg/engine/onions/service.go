package onions

import (
	"reflect"
	"time"

	"github.com/gookit/color"

	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/crypto/sha256"
	"github.com/indra-labs/indra/pkg/engine/coding"
	"github.com/indra-labs/indra/pkg/engine/magic"
	"github.com/indra-labs/indra/pkg/engine/node"
	"github.com/indra-labs/indra/pkg/engine/sess"
	"github.com/indra-labs/indra/pkg/engine/sessions"
	"github.com/indra-labs/indra/pkg/util/qu"
	"github.com/indra-labs/indra/pkg/util/slice"
	"github.com/indra-labs/indra/pkg/util/splice"
)

const (
	ServiceMagic = "serv"
	ServiceLen   = magic.Len +
		nonce.IDLen +
		crypto.PubKeyLen + 1 +
		splice.AddrLen +
		slice.Uint16Len +
		slice.Uint64Len +
		crypto.SigLen
)

type Service struct {
	ID        nonce.ID    // This ensures never a repeated signed message.
	Key       *crypto.Pub // Identity key.
	Port      uint16      // Well known port designating service protocol.
	RelayRate uint32      // Fee rate in mSat/Mb
	Expiry    time.Time
	Sig       crypto.SigBytes
}

func NewService(id nonce.ID, key *crypto.Prv, port uint16, relayRate uint32,
	expiry time.Time) (in *Service) {

	pk := crypto.DerivePub(key)
	s := splice.New(ServiceLen - magic.Len)
	s.ID(id).
		Pubkey(pk).
		Uint16(port).
		Uint32(relayRate).
		Uint64(uint64(expiry.UnixNano()))
	hash := sha256.Single(s.GetUntil(s.GetCursor()))
	var e error
	var sign crypto.SigBytes
	if sign, e = crypto.Sign(key, hash); fails(e) {
		return nil
	}
	in = &Service{
		ID:        id,
		Key:       pk,
		Port:      port,
		RelayRate: relayRate,
		Expiry:    expiry,
		Sig:       sign,
	}
	return
}

func (x *Service) Account(res *sess.Data, sm *sess.Manager,
	s *sessions.Data, last bool) (skip bool, sd *sessions.Data) {

	res.ID = x.ID
	return
}

func (x *Service) Decode(s *splice.Splice) (e error) {
	if e = magic.TooShort(s.Remaining(), ServiceLen-magic.Len,
		ServiceMagic); fails(e) {

		return
	}
	s.ReadID(&x.ID).
		ReadPubkey(&x.Key).
		ReadUint16(&x.Port).
		ReadUint32(&x.RelayRate).
		ReadTime(&x.Expiry).
		ReadSignature(&x.Sig)
	return
}

func (x *Service) Encode(s *splice.Splice) (e error) {
	log.T.S("encoding", reflect.TypeOf(x),
		x.ID, x.Expiry, x.Sig,
	)
	x.Splice(s.Magic(ServiceMagic))
	return
}

func (x *Service) GetOnion() interface{} { return x }

func (x *Service) Gossip(sm *sess.Manager, c qu.C) {
	log.D.F("propagating peer info for %s",
		x.Key.ToBased32Abbreviated())
	Gossip(x, sm, c)
	log.T.Ln("finished broadcasting peer info")
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

func (x *Service) Len() int      { return ServiceLen }
func (x *Service) Magic() string { return ServiceMagic }

func (x *Service) Splice(s *splice.Splice) {
	s.ID(x.ID).
		Pubkey(x.Key).
		Uint16(x.Port).
		Uint32(x.RelayRate).
		Uint64(uint64(x.Expiry.UnixNano())).
		Signature(x.Sig)
}

func (x *Service) Validate() bool {
	s := splice.New(ServiceLen - magic.Len)
	s.ID(x.ID).
		Pubkey(x.Key).
		Uint16(x.Port).
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

func (x *Service) Wrap(inner Onion) {}
func init()                         { Register(ServiceMagic, servGen) }
func servGen() coding.Codec { return &Service{} }
