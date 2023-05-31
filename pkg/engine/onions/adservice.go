package onions

import (
	"fmt"
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/crypto/sha256"
	"github.com/indra-labs/indra/pkg/engine/coding"
	"github.com/indra-labs/indra/pkg/engine/magic"
	"github.com/indra-labs/indra/pkg/engine/sess"
	"github.com/indra-labs/indra/pkg/util/qu"
	"github.com/indra-labs/indra/pkg/util/slice"
	"github.com/indra-labs/indra/pkg/util/splice"
)

const (
	ServiceAdMagic = "svad"
	ServiceAdLen   = magic.Len +
		nonce.IDLen +
		crypto.PubKeyLen +
		slice.Uint16Len +
		slice.Uint32Len +
		crypto.SigLen
)

// ServiceAd stores a specification for the fee rate and the service port, which
// must be a well known port to match with a type of service, eg 80 for web, 53
// for DNS, etc. These are also attached to the PeerAd entry via concatenating
// "/service/N" where N is the index of the entry. A zero value at an index
// signals to stop scanning for more subsequent values.
type ServiceAd struct {
	ID        nonce.ID    // To ensure no repeating message.
	Key       *crypto.Pub // Server offering service.
	Port      uint16
	RelayRate uint32
	Sig       crypto.SigBytes
}

var _ coding.Codec = &ServiceAd{}

func NewServiceAd(
	id nonce.ID,
	key *crypto.Prv,
	relayRate uint32,
	port uint16,
) (sv *ServiceAd) {

	s := splice.New(IntroLen)
	k := crypto.DerivePub(key)
	ServiceSplice(s, id, k, relayRate, port)
	hash := sha256.Single(s.GetUntil(s.GetCursor()))
	var e error
	var sign crypto.SigBytes
	if sign, e = crypto.Sign(key, hash); fails(e) {
		return nil
	}
	sv = &ServiceAd{
		ID:        id,
		Key:       k,
		RelayRate: relayRate,
		Port:      port,
		Sig:       sign,
	}
	return
}

func (x *ServiceAd) Decode(s *splice.Splice) (e error) {
	s.ReadID(&x.ID).
		ReadPubkey(&x.Key).
		ReadUint16(&x.Port).
		ReadUint32(&x.RelayRate).
		ReadSignature(&x.Sig)
	return
}

func (x *ServiceAd) Encode(s *splice.Splice) (e error) {
	x.Splice(s)
	return
}

func (x *ServiceAd) GetOnion() interface{} { return nil }

func (x *ServiceAd) Gossip(sm *sess.Manager, c qu.C) {
}

func (x *ServiceAd) Len() int { return ServiceAdLen }

func (x *ServiceAd) Magic() string { return "" }

func (x *ServiceAd) Sign(prv *crypto.Prv) (e error) {
	s := splice.New(x.Len())
	if e = x.Encode(s); fails(e) {
		return
	}
	var b []byte
	if b, e = prv.Sign(s.GetUntil(s.GetCursor())); fails(e) {
		return
	}
	if len(b) != crypto.SigLen {
		return fmt.Errorf("signature incorrect length, got %d expected %d",
			len(b), crypto.SigLen)
	}
	copy(x.Sig[:], b)
	return nil
}

func (x *ServiceAd) Splice(s *splice.Splice) {
	x.SpliceNoSig(s)
	s.Signature(x.Sig)
}

func (x *ServiceAd) SpliceNoSig(s *splice.Splice) {
	ServiceSplice(s, x.ID, x.Key, x.RelayRate, x.Port)
}

func (x *ServiceAd) Validate() (valid bool) {
	s := splice.New(IntroLen - magic.Len)
	x.SpliceNoSig(s)
	hash := sha256.Single(s.GetUntil(s.GetCursor()))
	key, e := x.Sig.Recover(hash)
	if fails(e) {
		return false
	}
	if key.Equals(x.Key) {
		return true
	}
	return false
}

func ServiceSplice(
	s *splice.Splice,
	id nonce.ID,
	key *crypto.Pub,
	relayRate uint32,
	port uint16,
) {

	s.Magic(ServiceAdMagic).
		ID(id).
		Pubkey(key).
		Uint16(port).
		Uint32(relayRate)
}

func init()                      { Register(ServiceAdMagic, serviceAdGen) }

func serviceAdGen() coding.Codec { return &ServiceAd{} }
