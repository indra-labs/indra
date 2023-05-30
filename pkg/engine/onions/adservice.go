package onions

import (
	"fmt"
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/crypto/sha256"
	"github.com/indra-labs/indra/pkg/engine/coding"
	"github.com/indra-labs/indra/pkg/engine/sess"
	"github.com/indra-labs/indra/pkg/engine/sessions"
	"github.com/indra-labs/indra/pkg/util/qu"
	"github.com/indra-labs/indra/pkg/util/slice"
	"github.com/indra-labs/indra/pkg/util/splice"
)

const (
	ServiceAdMagic = "intr"
	ServiceAdLen   = nonce.IDLen +
		2*slice.Uint16Len +
		slice.Uint32Len +
		crypto.SigLen
)

// ServiceAd stores a specification for the fee rate and the service port, which
// must be a well known port to match with a type of service, eg 80 for web, 53
// for DNS, etc. These are also attached to the Peer entry via concatenating
// "/service/N" where N is the index of the entry. A zero value at an index
// signals to stop scanning for more subsequent values.
type ServiceAd struct {
	ID        nonce.ID // To ensure no repeating message
	Index     uint16   // This is the index in the slice from Peer.
	Port      uint16
	RelayRate uint32
	Sig       crypto.SigBytes
}

func (x *ServiceAd) Account(res *sess.Data, sm *sess.Manager, s *sessions.Data, last bool) (skip bool, sd *sessions.Data) {
	return false, nil
}

func (x *ServiceAd) Decode(s *splice.Splice) (e error) {
	s.ReadID(&x.ID).
		ReadUint16(&x.Index).
		ReadUint16(&x.Port).
		ReadUint32(&x.RelayRate)
	return
}
func (x *ServiceAd) Encode(s *splice.Splice) (e error) {
	x.Splice(s.Magic(ServiceAdMagic))
	return
}

func (x *ServiceAd) GetOnion() interface{} { return nil }

func (x *ServiceAd) Gossip(sm *sess.Manager, c qu.C) {
}

func (x *ServiceAd) Handle(s *splice.Splice, p Onion, ni Ngin) (e error) {
	return nil
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
	s.ID(x.ID).
		Uint16(x.Index).
		Uint16(x.Port).
		Uint32(x.RelayRate)
}

func (x *ServiceAd) Validate(s *splice.Splice) (pub *crypto.Pub) {
	h := sha256.Single(s.GetRange(0, nonce.IDLen+2*slice.Uint16Len+
		slice.Uint64Len))
	var e error
	if pub, e = x.Sig.Recover(h); fails(e) {
	}
	return
}

func (x *ServiceAd) Wrap(inner Onion) {}

func init()                      { Register(ServiceAdMagic, serviceAdGen) }
func serviceAdGen() coding.Codec { return &Intro{} }
