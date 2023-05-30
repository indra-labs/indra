package onions

import (
	"fmt"
	"github.com/indra-labs/indra/pkg/crypto"
	"github.com/indra-labs/indra/pkg/crypto/nonce"
	"github.com/indra-labs/indra/pkg/crypto/sha256"
	"github.com/indra-labs/indra/pkg/util/slice"
	"github.com/indra-labs/indra/pkg/util/splice"
)

const ServiceAdLen = nonce.IDLen +
	2*slice.Uint16Len +
	slice.Uint32Len +
	crypto.SigLen

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

func (sv *ServiceAd) Decode(s *splice.Splice) (e error) {
	s.ReadID(&sv.ID).
		ReadUint16(&sv.Index).
		ReadUint16(&sv.Port).
		ReadUint32(&sv.RelayRate)
	return
}

func (sv *ServiceAd) Encode(s *splice.Splice) (e error) {
	s.ID(sv.ID).Uint16(sv.Index).Uint16(sv.Port).Uint32(sv.RelayRate)
	return
}

func (sv *ServiceAd) GetOnion() interface{} { return nil }

func (sv *ServiceAd) Len() int { return ServiceAdLen }

func (sv *ServiceAd) Magic() string { return "" }

func (sv *ServiceAd) Sign(prv *crypto.Prv) (e error) {
	s := splice.New(sv.Len())
	if e = sv.Encode(s); fails(e) {
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
	copy(sv.Sig[:], b)
	return nil
}

func (sv *ServiceAd) Validate(s *splice.Splice) (pub *crypto.Pub) {
	h := sha256.Single(s.GetRange(0, nonce.IDLen+2*slice.Uint16Len+
		slice.Uint64Len))
	var e error
	if pub, e = sv.Sig.Recover(h); fails(e) {
	}
	return
}
